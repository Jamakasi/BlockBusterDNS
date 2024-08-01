package dmatcher

import (
	"bufio"
	"errors"
	"io"
	"os"
	"strings"

	clog "github.com/coredns/coredns/plugin/pkg/log"
	dbc "github.com/patrickmn/go-cache"
)

type SMemDB struct {
	FilePath string
	memcache *dbc.Cache
	log      clog.P
}

func NewMemDB(filePath string, log clog.P) *SMemDB {
	ins := &SMemDB{
		FilePath: filePath,
		memcache: dbc.New(dbc.NoExpiration, dbc.NoExpiration),
		log:      log,
	}

	//c.Set("baz", 42, dbc.NoExpiration)
	ins.Load()
	return ins
}

// interface implement
func (t *SMemDB) AddDomain(val string) error {
	t.log.Debugf("add domain: %s", val)
	if strings.LastIndex(val, ".") != len(val)-1 {
		val = val + "."
	}
	if strings.HasPrefix(val, "*") {
		t.memcache.Set(val[2:], true, dbc.NoExpiration)
		t.log.Infof("add domain: %s. Is wildcard: %t", val[2:], true)
	} else {
		t.memcache.Set(val, false, dbc.NoExpiration)
		t.log.Infof("add domain: %s. Is wildcard: %t", val, false)
	}
	return nil
}
func (t *SMemDB) DelDomain(val string) error {
	t.log.Debugf("del domain: %s", val)
	if strings.HasPrefix(val, "*") {
		t.memcache.Delete(val[2:])
		t.log.Infof("add domain: %s. Is wildcard: %t", val[2:], true)
	} else {
		t.memcache.Delete(val)
		t.log.Infof("add domain: %s. Is wildcard: %t", val, false)
	}
	return nil
}
func (t *SMemDB) ContainDomain(val string) (bool, error) {
	domain := val
	for i := 0; i != -1; i = strings.Index(domain, ".") {
		if i != 0 {
			domain = domain[i+1:]
		}
		iswildcard, found := t.memcache.Get(domain)
		if found && i == 0 {
			t.log.Infof("match domain in zero iteration: %s. Is wildcard: %t", domain, iswildcard.(bool))
			return true, nil
		}
		if found && iswildcard.(bool) == true {
			t.log.Infof("match domain in %d iteration: %s, query domain %s. Is wildcard: %t", i, domain, val, iswildcard.(bool))
			return true, nil
		}
	}
	return false, nil
}
func (t *SMemDB) GetDomainList(q string) ([]string, error) {
	list := make([]string, 0)
	for key, val := range t.memcache.Items() {
		if strings.HasSuffix(key, q) {
			if val.Object.(bool) {
				list = append(list, "*."+key)
			} else {
				list = append(list, key)
			}
		}
	}
	return list, nil
}
func (t *SMemDB) Load() {
	file, err := os.Open(t.FilePath)
	if err != nil {
		t.log.Errorf("failed to load db from file: %s\n Error", t.FilePath, err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// optionally, resize scanner's capacity for lines over 64K, see next example
	count := 0
	for scanner.Scan() {
		//fmt.Printf("readed line %v %v\n", count, scanner.Text())
		count++
		t.AddDomain(scanner.Text())
	}
	t.log.Infof("load from db file %s. Total %d records", t.FilePath, count)
}
func (t *SMemDB) Save() {

	if _, err := os.Stat(t.FilePath); err == nil {
		os.Remove(t.FilePath)

	} else if errors.Is(err, os.ErrNotExist) {
		// path/to/whatever does *not* exist

	} else {
		// Schrodinger: file may or may not exist. See err for details.
		// Therefore, do *NOT* use !os.IsNotExist(err) to test for file existence

	}
	f, err := os.Create(t.FilePath)
	if err != nil {
		t.log.Errorf("failed save db to file: %s\n Error", t.FilePath, err)
		return
	}
	defer f.Close()
	list, _ := t.GetDomainList(".")
	for _, s := range list {
		_, err = io.WriteString(f, s+"\n")
		if err != nil {
			//log.Println("failed to save: " + err.Error())
		}
	}
	t.log.Infof("save to db file %s. Total %d records", t.FilePath, len(list))
}
