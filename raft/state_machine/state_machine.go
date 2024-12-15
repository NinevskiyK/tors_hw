package statemachine

import (
	"errors"
	"strings"
)

var kv_store map[string]string

func Init() {
	kv_store = make(map[string]string)
}

func Create(k string, v string) error {
	_, exists := kv_store[k]
	if exists {
		return errors.New("value already exists")
	}
	kv_store[k] = v
	return nil
}

func Update(k string, v string) error {
	_, exists := kv_store[k]
	if !exists {
		return errors.New("value doesnt exists")
	}
	kv_store[k] = v
	return nil
}

func Delete(k string) error {
	_, exists := kv_store[k]
	if !exists {
		return errors.New("value doesnt exists")
	}
	delete(kv_store, k)
	return nil
}

func Read(k string) (string, error) {
	_, exists := kv_store[k]
	if !exists {
		return "", errors.New("value doesnt exists")
	}
	return kv_store[k], nil
}

func CAS(k, body string) error {
	value, exists := kv_store[k]
	if !exists {
		return errors.New("value doesn't exists")
	}
	splitted := strings.Split(body, " ")
	if len(splitted) < 2 {
		return errors.New("wrong request")
	}
	old, new := splitted[0], splitted[1]
	if value == old {
		kv_store[k] = new
		return nil
	}
	return errors.New("wrong value")
}
