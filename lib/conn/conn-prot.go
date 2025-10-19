package conn

import (
	"bytes"
	"encoding/json"
	"math/big"
	"net"
	"strings"
	"time"

	"msh/lib/config"
	"msh/lib/errco"
	"msh/lib/model"
)

// buildMessage takes the request type and message to write to the client
func buildMessage(reqType int, message string) []byte {
	// mountHeader mounts the full header to a specified message
	var mountHeader = func(data []byte) []byte {
		//                  ┌--------------------full header--------------------┐
		// scheme:          [ sub-header1     | sub-header2 | sub-header3       | message   ]
		// bytes used:      [ 2               | 1           | 2                 | 0 - 16379 ]
		// value range:     [ 128 0 - 255 127 | 0           | 128 0 - 255 127	| --------- ]

		// addSubHeader mounts 1 sub-header to a specified message
		var addSubHeader = func(message []byte) []byte {
			//              ┌------sub-header1/3------┐
			// scheme:      [ firstByte | secondByte  | data ]
			// value range: [ 128 - 255 | 0 - 127     | ---- ]
			// it's a number composed of 2 digits in base-128 (firstByte is least significant byte)
			// sub-header represents the length of the following data

			firstByte := len(message)%128 + 128
			secondByte := float64(len(message) / 128)
			return append([]byte{byte(firstByte), byte(secondByte)}, message...)
		}

		// sub-header3 calculation
		data = addSubHeader(data)

		// sub-header2 calculation
		data = append([]byte{0}, data...)

		// sub-header1 calculation
		data = addSubHeader(data)

		return data
	}

	switch reqType {

	// send text to be shown in the loadscreen
	case errco.CLIENT_REQ_JOIN:
		messageStruct := &model.DataTxt{}
		messageStruct.Text = message

		dataTxtJSON, err := json.Marshal(messageStruct)
		if err != nil {
			// don't return error, just log a warning
			errco.NewLogln(errco.TYPE_WAR, errco.LVL_3, errco.ERROR_JSON_MARSHAL, err.Error())
			return nil
		}

		return mountHeader(dataTxtJSON)

	// send server info
	case errco.CLIENT_REQ_INFO:

		// "&" [\x26] is converted to "§" [\xc2\xa7]
		// this step is not strictly necessary if in msh-config is used the character "§"
		message = strings.ReplaceAll(message, "&", "§")

		// replace "\\n" with "\n" in case the new line was set as msh parameter
		message = strings.ReplaceAll(message, "\\n", "\n")

		messageStruct := &model.DataInfo{}
		messageStruct.Description.Text = message
		messageStruct.Players.Max = 0
		messageStruct.Players.Online = 0
		messageStruct.Version.Name = config.ConfigRuntime.Server.Version
		messageStruct.Version.Protocol = config.ConfigRuntime.Server.Protocol
		messageStruct.Favicon = "data:image/png;base64," + config.ServerIcon

		dataInfJSON, err := json.Marshal(messageStruct)
		if err != nil {
			// don't return error, just log a warning
			errco.NewLogln(errco.TYPE_WAR, errco.LVL_3, errco.ERROR_JSON_MARSHAL, err.Error())
			return nil
		}

		return mountHeader(dataInfJSON)

	default:
		return nil
	}
}

// getReqType returns the request packet, type (INFO or JOIN).
// Not player name as it's too difficult to extract.
func getReqType(clientConn net.Conn) ([]byte, int, *errco.MshLog) {
	var dataReqFull []byte

	data, logMsh := getClientPacket(clientConn)
	if logMsh != nil {
		return nil, errco.CLIENT_REQ_UNKN, logMsh.AddTrace()
	}

	dataReqFull = data

	// generate flags
	ProxyPortByt := big.NewInt(int64(config.ProxyPort)).Bytes() // calculates listen port in BigEndian bytes
	reqFlagInfo := append(ProxyPortByt, byte(1))              // flag contained in INFO request packet -> [99 211 1]
	reqFlagJoin := append(ProxyPortByt, byte(2))              // flag contained in JOIN request packet -> [99 211 2]

	// extract request type key byte
	reqTypeKeyByte := byte(0)
	if len(dataReqFull) > int(dataReqFull[0]) {
		reqTypeKeyByte = dataReqFull[int(dataReqFull[0])]
	}

	switch {
	case reqTypeKeyByte == byte(1) || bytes.Contains(dataReqFull, reqFlagInfo):
		// client is requesting server info
		// example: [ 16 0 244 5 9 49 50 55 46 48 46 48 46 49 99 211 1 1 0 ]
		//  ______________ case 1 _____________      ____________ case 2 ___________
		// [ 16 ... x x x (ProxyPortBytes) 1 1 0 ] or [ 16 ... x x x (ProxyPortBytes) 1 ]
		// [              ^-reqFlagInfo--^     ]    [              ^-reqFlagInfo--^ ]
		// [    ^-----------16 bytes-----^     ]    [    ^-----------16 bytes-----^ ]

		return dataReqFull, errco.CLIENT_REQ_INFO, nil

	case reqTypeKeyByte == byte(2) || bytes.Contains(dataReqFull, reqFlagJoin):
		// client is trying to join the server
		// example: [ 16 0 244 5 9 49 50 55 46 48 46 48 46 49 99 211 2 ]
		//  _______________________ case 1 ______________________      ________________________ case 2 ________________________
		// [ 16 ... x x x (ProxyPortBytes) 2 x ... x (player name) ] or [ 16 ... x x x (ProxyPortBytes) 2 ][ x ... x (player name) ]
		// [              ^-reqFlagJoin--^                       ]    [              ^-reqFlagJoin--^ ][                       ]
		// [    ^-----------16 bytes-----^                       ]    [    ^-----------16 bytes-----^ ][                       ]

		// after ms 1.19.3 not always there is a EOF after case 1/2
		// it's important to calculate if msh should read an other packet
		// bugfix #197
		switch {
		case len(dataReqFull) < int(dataReqFull[0])+1:
			// case unexpected: should not be possible
			return nil, errco.CLIENT_REQ_UNKN, errco.NewLog(errco.TYPE_WAR, errco.LVL_3, errco.ERROR_ANALYSIS, "unexpected data lenght")

		case len(dataReqFull) == int(dataReqFull[0])+1:
			// case 1: msh still has a packet to read
			data, logMsh = getClientPacket(clientConn)
			if logMsh != nil {
				return nil, errco.CLIENT_REQ_UNKN, logMsh.AddTrace() // return request unknown as the request failed
			}
			dataReqFull = append(dataReqFull, data...)

		case len(dataReqFull) > int(dataReqFull[0])+1:
			// case 2 (probably): no need to read more data from client
		}

		return dataReqFull, errco.CLIENT_REQ_JOIN, nil

	default:
		return nil, errco.CLIENT_REQ_UNKN, errco.NewLog(errco.TYPE_ERR, errco.LVL_3, errco.ERROR_CLIENT_REQ, "client request unknown (received: %v)", dataReqFull)
	}
}

// getPing performs msh PING response to the client PING request
// (must be performed after msh INFO response)
func getPing(clientConn net.Conn) *errco.MshLog {
	// read the first packet
	pingData, logMsh := getClientPacket(clientConn)
	if logMsh != nil {
		return logMsh.AddTrace()
	}

	switch {
	case bytes.HasPrefix(pingData, []byte{9, 1, 0, 0, 0, 0, 0}):
		// this is a normal ping
		// packet is [9 1 0 0 0 0 0 89 73 114]
		// using [9 1 0 0 0 0 0] as prefix is a conservative check
		// in case other "normal" pings are discovered it's possible to use a shorter/different prefix check

	case bytes.Equal(pingData, []byte{1, 0}):
		// packet is [1 0]
		// read the second packet
		pingData, logMsh = getClientPacket(clientConn)
		if logMsh != nil {
			return logMsh.AddTrace()
		}

	// bufix #202: don't use `bytes.Equal(pingData[:2], []byte{1, 0})`
	// (if pingData == [1] then pingData[:2] == [1 0])
	// this results in pingData[2:] -> slice bounds out of range [2:1]
	case bytes.HasPrefix(pingData, []byte{1, 0}):
		// packet is [1 0 9 1 0 0 0 0 0 89 73 114]
		// remove first 2 bytes: [1 0 9 1 0 0 0 0 0 89 73 114] -> [9 1 0 0 0 0 0 89 73 114]
		pingData = pingData[2:]

	default:
		return errco.NewLog(errco.TYPE_WAR, errco.LVL_3, errco.ERROR_PING_PACKET_UNKNOWN, "received unknown ping packet: %v", pingData)
	}

	// answer ping
	clientConn.Write(pingData)

	errco.NewLogln(errco.TYPE_BYT, errco.LVL_4, errco.ERROR_NIL, "%sproxy --> client%s: %v", errco.COLOR_PURPLE, errco.COLOR_RESET, pingData)

	return nil
}

// getClientPacket reads the client socket and returns only the bytes containing data
// clientConn connection should not be closed here (need to be closed in caller function).
func getClientPacket(clientConn net.Conn) ([]byte, *errco.MshLog) {
	buf := make([]byte, 1024)

	// set deadline to avoid hanging when client is not sending a packet that msh expects
	clientConn.SetDeadline(time.Now().Add(1 * time.Second))

	// read first packet
	dataLen, err := clientConn.Read(buf)
	if err != nil {
		return nil, errco.NewLog(errco.TYPE_ERR, errco.LVL_3, errco.ERROR_CLIENT_SOCKET_READ, err.Error())
	}

	errco.NewLogln(errco.TYPE_BYT, errco.LVL_4, errco.ERROR_NIL, "%sclient --> proxy%s: %v", errco.COLOR_PURPLE, errco.COLOR_RESET, buf[:dataLen])

	return buf[:dataLen], nil
}
