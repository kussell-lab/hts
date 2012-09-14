// Copyright ©2012 Dan Kortschak <dan.kortschak@adelaide.edu.au>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package bam

import (
	"errors"
	"fmt"
	"strings"
)

var (
	dupReference  = errors.New("bam: duplicate reference name")
	dupReadGroup  = errors.New("bam: duplicate read group name")
	dupProgram    = errors.New("bam: duplicate program name")
	usedReference = errors.New("bam: reference already used")
	usedReadGroup = errors.New("bam: read group already used")
	usedProgram   = errors.New("bam: program already used")
	dupRefLen     = errors.New("bam: repeated reference length")
	badLen        = errors.New("bam: reference length out of range")
)

type SortOrder int

const (
	UnknownOrder SortOrder = iota
	Unsorted
	QueryName
	Coordinate
)

var (
	sortOrder = []string{
		UnknownOrder: "unknown",
		Unsorted:     "unsorted",
		QueryName:    "queryname",
		Coordinate:   "coordinate",
	}
	sortOrderMap = map[string]SortOrder{
		"unknown":    UnknownOrder,
		"unsorted":   Unsorted,
		"queryname":  QueryName,
		"coordinate": Coordinate,
	}
)

func (so SortOrder) String() string {
	if so < Unsorted || so > Coordinate {
		return sortOrder[UnknownOrder]
	}
	return sortOrder[so]
}

type set map[string]int32

type Header struct {
	Version    string
	SortOrder  SortOrder
	refs       []*Reference
	rgs        []*ReadGroup
	progs      []*Program
	comments   []string
	seenRefs   set
	seenGroups set
	seenProgs  set
}

func NewHeader(text []byte, r []*Reference) (*Header, error) {
	var err error
	bh := &Header{refs: r, seenRefs: set{}, seenGroups: set{}}
	for i, r := range bh.refs {
		r.id = int32(i)
	}
	if text != nil {
		err = bh.parseHeader(text)
		if err != nil {
			return nil, err
		}
	}
	return bh, nil
}

func (bh *Header) String() string {
	var refs = make([]string, len(bh.refs))
	for i, r := range bh.refs {
		refs[i] = r.String()
	}
	if bh.Version != "" {
		return fmt.Sprintf("@HD\tVN:%s\tSO:%s\n%v\n",
			bh.Version,
			bh.SortOrder,
			strings.Trim(strings.Join(refs, "\n"), "[]"))
	}
	return strings.Trim(strings.Join(refs, "\n"), "[]")
}

func (bh *Header) Copy() *Header {
	c := &Header{
		Version:    bh.Version,
		SortOrder:  bh.SortOrder,
		comments:   append([]string(nil), bh.comments...),
		refs:       make([]*Reference, len(bh.refs)),
		rgs:        make([]*ReadGroup, len(bh.rgs)),
		progs:      make([]*Program, len(bh.progs)),
		seenRefs:   make(set, len(bh.seenRefs)),
		seenGroups: make(set, len(bh.seenGroups)),
		seenProgs:  make(set, len(bh.seenProgs)),
	}

	for i, r := range bh.refs {
		*c.refs[i] = *r
	}
	for i, r := range bh.rgs {
		*c.rgs[i] = *r
		c.rgs[i].flowOrder = append([]byte(nil), r.flowOrder...)
		c.rgs[i].keySeq = append([]byte(nil), r.keySeq...)
	}
	for i, p := range bh.progs {
		*c.progs[i] = *p
	}
	for k, v := range bh.seenRefs {
		c.seenRefs[k] = v
	}
	for k, v := range bh.seenGroups {
		c.seenGroups[k] = v
	}
	for k, v := range bh.seenProgs {
		c.seenProgs[k] = v
	}

	return c
}

func (bh *Header) Bytes() []byte {
	return []byte(bh.String())
}

func (bh *Header) Len() int {
	return len(bh.refs)
}

func (bh *Header) Refs() []*Reference {
	return bh.refs
}

func (bh *Header) RGs() []*ReadGroup {
	return bh.rgs
}

func (bh *Header) Progs() []*Program {
	return bh.progs
}

func (bh *Header) AddReference(r *Reference) error {
	if dupID, dup := bh.seenRefs[r.name]; dup {
		if er := bh.refs[dupID]; *er == *r {
			return nil
		} else if tr := (Reference{id: er.id, name: er.name, lRef: er.lRef}); *er != tr {
			return dupReference
		}
		bh.refs[dupID] = r
		return nil
	}
	if r.id >= 0 {
		return usedReference
	}
	r.id = int32(len(bh.refs))
	bh.refs = append(bh.refs, r)
	return nil
}

func (bh *Header) AddReadGroup(r *ReadGroup) error {
	if _, ok := bh.seenGroups[r.name]; ok {
		return dupReadGroup
	}
	if r.id >= 0 {
		return usedReadGroup
	}
	r.id = int32(len(bh.rgs))
	bh.rgs = append(bh.rgs, r)
	return nil
}

func (bh *Header) AddProgram(r *ReadGroup) error {
	if _, ok := bh.seenProgs[r.name]; ok {
		return dupProgram
	}
	if r.id >= 0 {
		return usedProgram
	}
	r.id = int32(len(bh.rgs))
	bh.rgs = append(bh.rgs, r)
	return nil
}
