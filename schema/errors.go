// Copyright (c) 2017,2019, AT&T Intellectual Property. All rights reserved
//
// Copyright (c) 2016-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/danos/mgmterror"
	"github.com/danos/utils/pathutil"
)

const (
	msgEmptyLeafValue = "Value found for empty leaf"
	msgMissingChild   = "Node requires a child"
	msgMissingKey     = "List entry is missing key"
	msgMissingValue   = "Node requires a value"
	msgSchemaMismatch = "Doesn't match schema"
	msgInvalidPath    = "Path is invalid"
)

func newOperationFailedtError(path []string, msg string) error {
	err := mgmterror.NewOperationFailedApplicationError()
	err.Path = pathutil.Pathstr(path)
	err.Message = msg
	return err
}

func newInvalidValueError(path []string, msg string) error {
	err := mgmterror.NewInvalidValueApplicationError()
	if len(path) > 0 {
		err.Path = pathutil.Pathstr(path)
	}
	err.Message = msg
	return err
}

func NewMissingChildError(path []string) error {
	e := mgmterror.NewMissingElementApplicationError("<any child>")
	e.Path = pathutil.Pathstr(path)
	e.Message = msgMissingChild
	return e
}

func NewMissingValueError(path []string) error {
	return newInvalidValueError(path, msgMissingValue)
}

func NewEmptyLeafValueError(name string, path []string) error {
	e := mgmterror.NewUnknownElementApplicationError(name)
	e.Path = pathutil.Pathstr(path)
	e.Message = msgEmptyLeafValue
	return e
}

func NewMissingKeyError(path []string) error {
	e := mgmterror.NewOperationFailedApplicationError()
	e.Path = pathutil.Pathstr(path)
	e.Message = msgMissingKey
	return e
}

func NewSchemaMismatchError(name string, path []string) error {
	e := mgmterror.NewUnknownElementApplicationError(name)
	e.Path = pathutil.Pathstr(path)
	e.Message = msgSchemaMismatch
	return e
}

func NewInvalidPathError(path []string) error {
	switch len(path) {
	case 0:
		e := mgmterror.NewOperationFailedApplicationError()
		e.Message = msgInvalidPath
		return e
	case 1:
		e := mgmterror.NewUnknownElementApplicationError(path[0])
		e.Message = msgInvalidPath
		return e
	}
	e := mgmterror.NewUnknownElementApplicationError(path[len(path)-1])
	e.Path = pathutil.Pathstr(path[:len(path)-1])
	e.Message = msgInvalidPath
	return e
}

// Preferred to NewInvalidPathError, generating consistent error type
// and consistent path style (split between path and info fields).  Legacy
// one kept to avoid having to play with opd errors.
func NewPathInvalidError(path []string, invalidElem string) error {
	e := mgmterror.NewUnknownElementApplicationError(invalidElem)
	e.Message = msgInvalidPath
	e.Path = pathutil.Pathstr(path)
	return e
}

func NewNodeNotExistsError(path []string) error {
	e := mgmterror.NewDataMissingError()
	e.Path = pathutil.Pathstr(path)
	e.Message = "Node does not exist"
	return e
}

func NewNodeExistsError(path []string) error {
	e := mgmterror.NewDataExistsError()
	e.Path = pathutil.Pathstr(path)
	e.Message = "Node exists"
	return e
}
