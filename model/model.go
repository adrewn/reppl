package model

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	"github.com/ugorji/go/codec"
	rdef "go.polydawn.net/repeatr/api/def"
)

type Project struct {
	Tags       map[string]ReleaseRecord   // map tag->{ware,backstory}
	RunRecords map[string]*rdef.RunRecord // map rrhid->rr
	Memos      map[string]string          // index frmhid->rrhid
}

type ReleaseRecord struct {
	Ware         rdef.Ware
	RunRecordHID string // blank if a tag was manual
}

func (p *Project) Init() {
	p.Tags = make(map[string]ReleaseRecord)
	p.RunRecords = make(map[string]*rdef.RunRecord)
	p.Memos = make(map[string]string)
}

func (p *Project) WriteFile(filename string) {
	f, err := os.Create(filename)
	if err != nil {
		panic("error opening project file")
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	defer w.Flush()

	enc := codec.NewEncoder(w, &codec.JsonHandle{})
	err = enc.Encode(p)
	if err != nil {
		panic("could not write project file")
	}
	w.Write([]byte{'\n'})
}

func FromFile(filename string) Project {
	f, err := os.Open(filename)
	if err != nil {
		panic("error opening project file")
	}
	defer f.Close()

	r := bufio.NewReader(f)
	p := Project{}
	dec := codec.NewDecoder(r, &codec.JsonHandle{})
	err = dec.Decode(&p)
	if err != nil {
		panic("error reading project file")
	}
	return p
}

func (p *Project) PutManualTag(tag string, ware rdef.Ware) {
	_, hadPrev := p.Tags[tag]
	p.Tags[tag] = ReleaseRecord{ware, ""}
	if hadPrev {
		p.retainFilter()
	}
}

func (p *Project) DeleteTag(tag string) {
	_, hadPrev := p.Tags[tag]
	if hadPrev {
		delete(p.Tags, tag)
		p.retainFilter()
	}
}

func (p *Project) GetWareByTag(tag string) (rdef.Ware, error) {
	_, exists := p.Tags[tag]
	if exists {
		return p.Tags[tag].Ware, nil
	} else {
		return rdef.Ware{}, errors.New("not found")
	}
}

func (p *Project) PutResult(tag string, resultName string, rr *rdef.RunRecord) {
	p.Tags[tag] = ReleaseRecord{rr.Results[resultName].Ware, rr.HID}
	p.RunRecords[rr.HID] = rr
	p.Memos[rr.FormulaHID] = rr.HID
	p.retainFilter()
}

func (p *Project) retainFilter() {
	// "Sweep".  (The `Tags` map is the marks.)
	oldRunRecords := p.RunRecords
	p.RunRecords = make(map[string]*rdef.RunRecord)
	p.Memos = make(map[string]string)
	// Rebuild `RunRecords` by whitelisting prev values still ref'd by `Tags`.
	for tag, release := range p.Tags {
		if release.RunRecordHID == "" {
			continue // skip.  it's just a fiat release; doesn't ref anything.
		}
		runRecord, ok := oldRunRecords[release.RunRecordHID]
		if !ok {
			panic(fmt.Errorf("db integrity violation: dangling runrecord -- release %q points to %q", tag, release.RunRecordHID))
		}
		p.RunRecords[release.RunRecordHID] = runRecord
	}
	// Rebuild `Memos` index from `RunRecords`.
	for _, runRecord := range p.RunRecords {
		p.Memos[runRecord.FormulaHID] = runRecord.HID
	}
}
