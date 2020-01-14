/* *********************************************************************
 * This Source Code Form is copyright of 51Degrees Mobile Experts Limited.
 * Copyright 2017 51Degrees Mobile Experts Limited, 5 Charlotte Close,
 * Caversham, Reading, Berkshire, United Kingdom RG4 7BY
 *
 * This Source Code Form is the subject of the following patents and patent
 * applications, owned by 51Degrees Mobile Experts Limited of 5 Charlotte
 * Close, Caversham, Reading, Berkshire, United Kingdom RG4 7BY:
 * European Patent No. 2871816;
 * European Patent Application No. 17184134.9;
 * United States Patent Nos. 9,332,086 and 9,350,823; and
 * United States Patent Application No. 15/686,066.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0.
 *
 * If a copy of the MPL was not distributed with this file, You can obtain
 * one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, v. 2.0.
 ********************************************************************** */

// Package dd 51Degrees device detection package. Implements both pattern and
// hash device detection in Go. Links to C repositories for specific
// implementation.
package foddd

// #cgo LDFLAGS: -Ldevice-detection-cxx/lib -lfiftyone-pattern-c -lfiftyone-device-detection-c -lfiftyone-common-c
// #include <stdio.h>
// #include <stdlib.h>
// #include <string.h>
// #include "device-detection-cxx/src/pattern/pattern.h"
// #include "device-detection-cxx/src/pattern/fiftyone.h"
// #include "device-detection-cxx/src/fiftyone.h"
// void clearException(Exception *ex) { ex->status = NOT_SET; }
// StatusCode getExceptionOkay() { return NOT_SET; }
import "C"

import (
	"fmt"
	"unsafe"
)

// Error encapsulates a C exception for return to consumers of the package.
type Error struct {
	file     string
	function string
	line     int
	message  string
}

// manager contains references to all the items needed to perform device
// detection
type Manager struct {
	manager C.ResourceManager
	ex      C.Exception
	results *C.ResultsPattern
}

func CreateError(ex C.Exception, dataFilePath *C.char) *Error {
	var error *Error
	if ex.status != C.getExceptionOkay() {
		message := C.StatusGetMessage(ex.status, dataFilePath)
		defer C.free(unsafe.Pointer(message))
		error = &Error{
			C.GoString(ex.file),
			C.GoString(ex._func),
			int(ex.line),
			C.GoString(message),
		}
	}
	return error
}

func printError(error *Error) {
	fmt.Println(error.message)
	fmt.Println("File:", error.file)
	fmt.Println("Func:", error.function)
	fmt.Println("Line:", error.line)
}

func getPropertyValueAsString(
	results *C.ResultsPattern,
	propertyName string) string {
	seperator := C.char(',')
	var ex C.Exception
	C.clearException(&ex)
	cPropertyName := C.CString(propertyName)
	defer C.free(unsafe.Pointer(cPropertyName))
	var value = [1024]C.char{}
	C.ResultsPatternGetValuesString(
		results,
		cPropertyName,
		&value[0],
		C.size_t(len(value)),
		&seperator,
		&ex)
	return C.GoString(&value[0])
}

// Get returns the property value from a previous call to Process.
func (manager *Manager) Get(property string) string {
	var value = getPropertyValueAsString(manager.results, property)
	var error = CreateError(manager.ex, nil)
	if error != nil {
		panic(error)
	}
	return value
}

// Why is this returning the same manager?
// Process sets the results for the User-Agent provided.
func (manager *Manager) ProcessFromEvidence(evidence *Evidence) *Manager {
	C.ResultsPatternFromEvidence(
		manager.results,
		evidence.e,
		&manager.ex)
	var error = CreateError(manager.ex, nil)
	if error != nil {
		panic(error)
	}
	return manager
}

// Why is this returning the same manager?
// Process sets the results for the User-Agent provided.
func (manager *Manager) Process(userAgent string) *Manager {
	var cuserAgent = C.CString(userAgent)
	C.ResultsPatternFromUserAgent(
		manager.results,
		cuserAgent,
		C.strlen(cuserAgent),
		&manager.ex)
	var error = CreateError(manager.ex, nil)
	if error != nil {
		panic(error)
	}
	return manager
}

// Free the manager and all its resources.
func (manager *Manager) Free() {
	C.ResultsPatternFree(manager.results)
	C.ResourceManagerFree(&manager.manager)
}

// Create builds a new device detection manager. Must call Free on the result.
func Create(dataFilePath string, properties string) *Manager {
	var manager Manager

	// Convert the data file path into a C char* so that it can be passed
	// to the initialise method.
	cdataFilePath := C.CString(dataFilePath)
	defer C.free(unsafe.Pointer(cdataFilePath))

	// Convert the properties to a C char * so that it can be passed to the
	// initialise method.
	var propertiesRequired C.PropertiesRequired
	propertiesRequired.string = C.CString("ScreenPixelsWidth,HardwareModel,IsMobile,BrowserName")
	defer C.free(unsafe.Pointer(propertiesRequired.string))

	// Create a configuration to control device detection.
	var config = C.PatternBalancedConfig

	// Set the exception to the defaults.
	C.clearException(&manager.ex)

	var status = C.PatternInitManagerFromFile(
		&manager.manager,
		&config,
		&propertiesRequired,
		cdataFilePath,
		&manager.ex)

	if status == 0 {
		manager.results = C.ResultsPatternCreate(&manager.manager, 1, 0)
	}

	var error = CreateError(manager.ex, nil)
	if error != nil {
		panic(error)
	}

	return &manager
}

var FIFTYONE_DEGREES_EVIDENCE_HTTP_HEADER_STRING = C.FIFTYONE_DEGREES_EVIDENCE_HTTP_HEADER_STRING // An HTTP header
var FIFTYONE_DEGREES_EVIDENCE_HTTP_HEADER_IP_ADDRESSES = C.FIFTYONE_DEGREES_EVIDENCE_HTTP_HEADER_IP_ADDRESSES // < A list of IP addresses as a string to be parsed into a IP addresses collection.
var FIFTYONE_DEGREES_EVIDENCE_SERVER = C.FIFTYONE_DEGREES_EVIDENCE_SERVER // A server value e.g. client IP */
var FIFTYONE_DEGREES_EVIDENCE_QUERY = C.FIFTYONE_DEGREES_EVIDENCE_QUERY /**< A query string parameter */
var FIFTYONE_DEGREES_EVIDENCE_COOKIE = C.FIFTYONE_DEGREES_EVIDENCE_COOKIE /**< A cookie value */
var FIFTYONE_DEGREES_EVIDENCE_IGNORE = C.FIFTYONE_DEGREES_EVIDENCE_IGNORE /**< The evidence is invalid and should be ignored */

// mirrors the evidence.c file
type Evidence struct {
	capacity uint
	e *C.EvidenceKeyValuePairArray
}

func NewEvidence(size uint) Evidence {
	//EvidenceCreate
	evidence := Evidence{capacity: size}
	sizec := C.uint(size)
	evidence.e = C.EvidenceCreate(sizec)
	return evidence
}

func (e *Evidence) Free() {
	C.EvidenceFree(e.e)
	e = nil
}

func (e *Evidence) AddString(prefix string, field string, originalValue string) {

	emp := prefix + "." + field
	empC := C.CString(emp)
	defer C.free(unsafe.Pointer(empC))
	result := C.fiftyoneDegreesEvidenceMapPrefix(empC)
	enum := result.prefixEnum

	fieldC := C.CString(field)
	valueC := C.CString(originalValue)
	//defer C.free(unsafe.Pointer(valueC))

	C.EvidenceAddString(
		e.e,
		enum,
		fieldC,
		valueC)

}

func (evidence *C.EvidenceKeyValuePairArray) AddEvidence(key string, value string) *C.EvidenceKeyValuePairArray {
	var prefix = C.EvidenceMapPrefix(C.CString("header.user-agent"))
	C.EvidenceAddString(
		evidence,
		prefix.prefixEnum,
		C.CString(key),
		C.CString(value))
	return evidence
}