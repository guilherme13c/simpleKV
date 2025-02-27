package store

import (
	"encoding/gob"
	"fmt"
	"os"
)

func (s *store) saveSnapshot() error {
	file, err := os.Create(s.persistenceFile)
	if err != nil {
		return fmt.Errorf("could not create persistence file: %v", err)
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	err = encoder.Encode(s)
	if err != nil {
		return fmt.Errorf("could not encode store data: %v", err)
	}

	return nil
}

func (s *store) loadSnapshot() error {
	file, err := os.Open(s.persistenceFile)
	if err != nil {
		return fmt.Errorf("could not open persistence file: %v", err)
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	err = decoder.Decode(s)
	if err != nil {
		return fmt.Errorf("could not decode store data: %v", err)
	}

	return nil
}
