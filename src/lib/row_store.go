package pcs

import "log"
import "fmt"
import "path"
import "bytes"
import "encoding/gob"
import "io/ioutil"
import "os"

type RowSavedInt struct {
	Name  int16
	Value int64
}

type RowSavedStr struct {
	Name  int16
	Value string
}

type RowSavedSet struct {
	Name  int16
	Value []int32
}

type SavedRecord struct {
	Ints []RowSavedInt
	Strs []RowSavedStr
	Sets []RowSavedSet
}

func (s SavedRecord) toRecord(t *Table) *Record {
	r := Record{}
	r.Ints = IntArr{}
	r.Strs = StrArr{}
	r.Sets = SetArr{}

	b := t.LastBlock
	b.table = t
	r.block = &b

	for _, v := range s.Ints {
		r.ResizeFields(v.Name)
		r.Populated[v.Name] = INT_VAL
		r.Ints[v.Name] = IntField(v.Value)
		t.update_int_info(v.Name, v.Value)
	}

	for _, v := range s.Strs {
		r.ResizeFields(v.Name)
		r.AddStrField(t.get_string_for_key(int(v.Name)), v.Value)
	}

	for _, v := range s.Sets {
		r.ResizeFields(v.Name)
		r.Sets[v.Name] = v.Value
	}

	return &r
}

func (r Record) toSavedRecord() *SavedRecord {
	s := SavedRecord{}
	for k, v := range r.Ints {
		if r.Populated[k] == INT_VAL {
			s.Ints = append(s.Ints, RowSavedInt{int16(k), int64(v)})
		}
	}

	for k, v := range r.Strs {
		if r.Populated[k] == STR_VAL {
			col := r.block.getColumnInfo(int16(k))
			str_val := col.get_string_for_val(int32(v))
			s.Strs = append(s.Strs, RowSavedStr{int16(k), str_val})
		}
	}

	for k, v := range r.Sets {
		if r.Populated[k] == SET_VAL {
			s.Sets = append(s.Sets, RowSavedSet{int16(k), v})
		}
	}

	return &s

}

type SavedRecords struct {
	RecordList []*SavedRecord
}

func (t *Table) LoadSavedRecordsFromLog(filename string) []*SavedRecord {
	log.Println("LOADING RECORDS FROM LOG", filename)
	var marshalled_records []*SavedRecord

	file, err := os.Open(filename)

	if err != nil {
		log.Println("ERROR OPENING FILE", filename, err)
	}

	// Create an encoder and send a value.
	enc := gob.NewDecoder(file)
	err = enc.Decode(&marshalled_records)

	if err != nil {
		log.Println("ERROR LOADING INGESTION LOG", err)
	}

	return marshalled_records
}

func (t *Table) LoadRecordsFromLog(filename string) RecordList {
	log.Println("LOADING RECORDS FROM LOG", filename)
	var marshalled_records []*SavedRecord

	file, err := os.Open(filename)

	if err != nil {
		log.Println("ERROR OPENING FILE", filename, err)
	}

	// Create an encoder and send a value.
	enc := gob.NewDecoder(file)
	err = enc.Decode(&marshalled_records)

	if err != nil {
		log.Println("ERROR LOADING INGESTION LOG", err)
	}

	ret := make(RecordList, len(marshalled_records))

	for i, r := range marshalled_records {
		ret[i] = r.toRecord(t)
	}

	return ret
}

func (t *Table) AppendRecordsToLog(records RecordList, blockname string) {
	if len(records) == 0 {
		return
	}

	ingestdir := path.Join(*f_DIR, t.Name, "ingest")
	digestdir := path.Join(*f_DIR, t.Name, "digest")
	stomachedir := path.Join(*f_DIR, t.Name, "stomache")

	os.MkdirAll(digestdir, 0777)
	os.MkdirAll(ingestdir, 0777)
	os.MkdirAll(stomachedir, 0777)

	w, err := ioutil.TempFile(ingestdir, fmt.Sprintf("%s_", blockname))

	marshalled_records := make([]*SavedRecord, len(records))
	for i, r := range records {
		marshalled_records[i] = r.toSavedRecord()
	}

	var network bytes.Buffer // Stand-in for the network.

	log.Println("SAVING RECORDS", len(marshalled_records), "TO INGESTION LOG")

	// Create an encoder and send a value.
	enc := gob.NewEncoder(&network)
	err = enc.Encode(marshalled_records)

	if err != nil {
		log.Fatal("encode:", err)
	}

	filename := fmt.Sprintf("%s.db", w.Name())
	log.Println("NAME", w.Name())

	log.Println("SERIALIZED INTO LOG", filename, network.Len(), "BYTES", "( PER RECORD", network.Len()/len(marshalled_records), ")")

	network.WriteTo(w)
	os.Rename(w.Name(), filename)

}
