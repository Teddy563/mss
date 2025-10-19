package config

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/denisbrodbeck/machineid"

	"msh/lib/errco"
	"msh/lib/model"
	"msh/lib/opsys"
	"msh/lib/utility"
)

const instanceFile string = "mineplus.instance"
const CFLAG string = "/*\\"

type MshInstanceV model.MshInstanceV
type MshInstanceV0 model.MshInstanceV0

// MshID returns msh id. A new istance is created if not healthy/not existent.
func MshID() string {
	// if mineplus instance does not exist, generate a new one
	_, err := os.Stat(instanceFile)
	if errors.Is(err, os.ErrNotExist) {
		errco.NewLogln(errco.TYPE_WAR, errco.LVL_3, errco.ERROR_CONFIG_MSHID, "mineplus instance file does not exist")
		return newMshInstance("")
	}

	errco.NewLogln(errco.TYPE_INF, errco.LVL_3, errco.ERROR_NIL, "mineplus instance file exists")

	// read from file
	instanceData, err := os.ReadFile(instanceFile)
	if err != nil {
		errco.NewLogln(errco.TYPE_WAR, errco.LVL_3, errco.ERROR_CONFIG_MSHID, "mineplus instance file can't be read")
		return newMshInstance("")
	}

	// replace NULL char with CFLAG to prevent JSON format error and wrong health check
	instanceData = bytes.ReplaceAll(instanceData, []byte{0}, []byte(CFLAG))

	// extract mineplus instance version
	var iv *MshInstanceV = &MshInstanceV{}
	err = json.Unmarshal(instanceData, iv)
	if err != nil {
		errco.NewLogln(errco.TYPE_WAR, errco.LVL_3, errco.ERROR_CONFIG_MSHID, "mineplus instance file does not contain version or not json formatted")
		return newMshInstance("")
	}

	switch iv.V {
	case 0:
		var i *MshInstanceV0 = &MshInstanceV0{}

		// unmarshal mineplus.instance file data into instance struct
		err = json.Unmarshal(instanceData, i)
		if err != nil {
			errco.NewLogln(errco.TYPE_WAR, errco.LVL_3, errco.ERROR_CONFIG_MSHID, "mineplus instance file not json formatted")
			return newMshInstance("")
		}

		// mineplus instance health check
		if !i.okV0() {
			errco.NewLogln(errco.TYPE_WAR, errco.LVL_3, errco.ERROR_CONFIG_MSHID, "mineplus instance loaded is corrupted")
			return newMshInstance("")
		}

		errco.NewLogln(errco.TYPE_INF, errco.LVL_3, errco.ERROR_NIL, "mineplus instance loaded is healthy")

		return i.MshId
		// when mineplus instance version is upgraded, above line will be replaced by:
		// return newMshInstance(i.MshId)

	default:
		// mineplus instance version is unsupported, generate a new instance
		errco.NewLogln(errco.TYPE_WAR, errco.LVL_3, errco.ERROR_CONFIG_MSHID, "mineplus instance loaded is unsupported")

		return newMshInstance("")
	}
}

// newMshInstance generates a new instance file and returns a new mshid
func newMshInstance(mshIDrecord string) string {
	var i *MshInstanceV0 = &MshInstanceV0{}

	errco.NewLogln(errco.TYPE_INF, errco.LVL_3, errco.ERROR_NIL, "generating new mineplus instance")

	// touch instance file (to know in advance file id)
	f, err := os.Create(instanceFile)
	if err != nil {
		log.Fatalln(err.Error())
	}
	_ = f.Close()

	// generate instance parameters
	i.V = 0                                   // set instance file version
	i.CFlag = CFLAG                           // set copy flag to CFLAG
	i.MId, err = machineid.ProtectedID("mineplus") // get machine id
	if err != nil {
		errco.NewLogln(errco.TYPE_WAR, errco.LVL_3, errco.ERROR_CONFIG_MSHID, err.Error())
	}
	i.HostName, err = os.Hostname() // get instance hostname
	if err != nil {
		errco.NewLogln(errco.TYPE_WAR, errco.LVL_3, errco.ERROR_CONFIG_MSHID, err.Error())
	}
	i.FId, err = opsys.FileId(instanceFile) // get instance file id
	if err != nil {
		errco.NewLogln(errco.TYPE_WAR, errco.LVL_3, errco.ERROR_CONFIG_MSHID, err.Error())
	}
	// try to use mshID old record
	i.MshId = mshIDrecord
	if utility.Entropy(i.MshId) < 150 {
		// old mshID entropy is too low: generate new mshid
		i.MshId = genMshId()
	}

	// generate mineplus instance checksum
	i.CheckSum = i.calcCheckSumV0()

	// marshal instance to bytes
	instanceData, err := json.Marshal(i)
	if err != nil {
		log.Fatalln(err.Error())
	}

	// replace CFLAG with NULL char to prevent accidental copy of mineplus.instance
	instanceData = bytes.ReplaceAll(instanceData, []byte(CFLAG), []byte{0})

	// write to instance file
	err = os.WriteFile(instanceFile, instanceData, 0644)
	if err != nil {
		log.Fatalln(err.Error())
	}

	// instance health check at birth
	if !i.okV0() {
		errco.NewLogln(errco.TYPE_WAR, errco.LVL_3, errco.ERROR_CONFIG_MSHID, "generated mineplus instance is corrupted")
		return newMshInstance(mshIDrecord)
	}

	errco.NewLogln(errco.TYPE_INF, errco.LVL_3, errco.ERROR_NIL, "generated mineplus instance is healthy")

	return i.MshId
}

// ok verify that mineplus instance V0 is healthy
func (i *MshInstanceV0) okV0() bool {
	// check that instance exists
	if i == nil {
		errco.NewLogln(errco.TYPE_ERR, errco.LVL_3, errco.ERROR_CONFIG_MSHID, "mineplus instance struct not loaded")
		return false
	}

	// check checksum
	Checksum := i.calcCheckSumV0()
	if i.CheckSum != Checksum {
		errco.NewLogln(errco.TYPE_WAR, errco.LVL_3, errco.ERROR_CONFIG_MSHID,
			"mineplus instance verification: wrong checksum"+"\n"+
				"\tinst checksum "+i.CheckSum+"\n"+
				"\tfile checksum "+Checksum)
		return false
	}

	// check machine id
	MId, err := machineid.ProtectedID("mineplus") // get machine id
	if err != nil {
		errco.NewLogln(errco.TYPE_ERR, errco.LVL_3, errco.ERROR_CONFIG_MSHID, err.Error())
	}
	if i.MId != MId {
		errco.NewLogln(errco.TYPE_WAR, errco.LVL_3, errco.ERROR_CONFIG_MSHID,
			"mineplus instance verification: wrong machine id"+"\n"+
				"\tinst checksum "+i.MId+"\n"+
				"\tfile checksum "+MId)
		return false
	}

	// check hostname
	HostName, err := os.Hostname()
	if err != nil {
		errco.NewLogln(errco.TYPE_ERR, errco.LVL_3, errco.ERROR_CONFIG_MSHID, err.Error())
	}
	if i.HostName != HostName {
		errco.NewLogln(errco.TYPE_WAR, errco.LVL_3, errco.ERROR_CONFIG_MSHID,
			"mineplus instance verification: wrong hostname"+"\n"+
				"\tinst checksum "+i.HostName+"\n"+
				"\tfile checksum "+HostName)
		return false
	}

	// check file id
	FId, err := opsys.FileId(instanceFile)
	if err != nil {
		errco.NewLogln(errco.TYPE_ERR, errco.LVL_3, errco.ERROR_CONFIG_MSHID, err.Error())
	}
	if i.FId != FId {
		errco.NewLogln(errco.TYPE_WAR, errco.LVL_3, errco.ERROR_CONFIG_MSHID,
			"mineplus instance verification: wrong file id"+"\n"+
				"\tinst checksum "+strconv.FormatUint(i.FId, 10)+"\n"+
				"\tfile checksum "+strconv.FormatUint(FId, 10))
		return false
	}

	return true
}

// calcCheckSum calculates mineplus instance V0 checksum.
// CheckSum instance parameter is excluded from computation.
func (i *MshInstanceV0) calcCheckSumV0() string {
	hasher := sha1.New()

	v := reflect.ValueOf(*i)
	t := v.Type()
	o := ""
	for i := 0; i < v.NumField(); i++ {
		// skip CheckSum field as we are calculating it
		if t.Field(i).Name == "CheckSum" {
			continue
		}
		o += fmt.Sprintf("%v", v.Field(i))
	}

	hasher.Write([]byte(o))
	return hex.EncodeToString(hasher.Sum(nil))
}

// genMshId generates a new mshID with Shannon entropy above 150 bits
func genMshId() string {
	rand.Seed(time.Now().UnixNano())
	mshID := ""

	// mshID must have a Shannon entropy of more than 150 bits
	for utility.Entropy(mshID) <= 150 {
		key := make([]byte, 64)
		_, _ = rand.Read(key) // returned error is always nil
		hasher := sha1.New()
		hasher.Write(key)
		mshID = hex.EncodeToString(hasher.Sum(nil))
	}

	return mshID
}
