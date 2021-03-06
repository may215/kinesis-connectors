package connector

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/s3"
)

// S3Emitter is an implementation of Emitter used to store files from a Kinesis stream in S3.
//
// The use of  this struct requires the configuration of an S3 bucket/endpoint. When the buffer is full, this
// struct's Emit method adds the contents of the buffer to S3 as one file. The filename is generated
// from the first and last sequence numbers of the records contained in that file separated by a
// dash. This struct requires the configuration of an S3 bucket and endpoint.
type S3Emitter struct {
	S3Bucket string
}

// S3FileName generates a file name based on the First and Last sequence numbers from the buffer. The current
// UTC date (YYYY-MM-DD) is base of the path to logically group days of batches.
func (e S3Emitter) S3FileName(firstSeq string, lastSeq string) string {
	date := time.Now().UTC().Format("2006/01/02")
	return fmt.Sprintf("%v/%v-%v", date, firstSeq, lastSeq)
}

// Emit is invoked when the buffer is full. This method emits the set of filtered records.
func (e S3Emitter) Emit(b Buffer, t Transformer) {
	auth, _ := aws.EnvAuth()
	s3Con := s3.New(auth, aws.USEast)
	bucket := s3Con.Bucket(e.S3Bucket)
	s3File := e.S3FileName(b.FirstSequenceNumber(), b.LastSequenceNumber())

	var buffer bytes.Buffer

	for _, r := range b.Records() {
		var s = t.FromRecord(r)
		buffer.Write(s)
	}

	err := bucket.Put(s3File, buffer.Bytes(), "text/plain", s3.Private, s3.Options{})

	if err != nil {
		log.Printf("S3Put ERROR: %v\n", err)
	} else {
		log.Printf("[%v] records emitted to [%s]\n", b.NumRecordsInBuffer(), e.S3Bucket)
	}
}
