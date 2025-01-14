package services

import (
	"encoding/json"
	"errors"
	"time"

	"rpi-heating-system/lib/model"

	"github.com/boltdb/bolt"
)

const bucketName = "Thermostats"

type ThermostatStorageService struct {
	db *bolt.DB
}

func NewThermostatStorageService(dbPath string) (*ThermostatStorageService, error) {
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}

	// Ensure the bucket exists
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		return err
	})
	if err != nil {
		return nil, err
	}

	return &ThermostatStorageService{db: db}, nil
}

func (s *ThermostatStorageService) Save(thermostat *model.Thermostat) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		data, err := json.Marshal(thermostat)
		if err != nil {
			return err
		}
		return b.Put([]byte(thermostat.ID), data)
	})
}

func (s *ThermostatStorageService) GetAll() ([]model.Thermostat, error) {
	var thermostats []model.Thermostat
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		return b.ForEach(func(k, v []byte) error {
			var thermostat model.Thermostat
			if err := json.Unmarshal(v, &thermostat); err != nil {
				return err
			}
			thermostats = append(thermostats, thermostat)
			return nil
		})
	})
	return thermostats, err
}

func (s *ThermostatStorageService) GetByID(id string) (*model.Thermostat, error) {
	var thermostat model.Thermostat
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		data := b.Get([]byte(id))
		if data == nil {
			return errors.New("thermostat not found")
		}
		return json.Unmarshal(data, &thermostat)
	})
	if err != nil {
		return nil, err
	}
	return &thermostat, nil
}
