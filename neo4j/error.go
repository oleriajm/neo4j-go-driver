/*
 * Copyright (c) 2002-2018 "Neo4j,"
 * Neo4j Sweden AB [http://neo4j.com]
 *
 * This file is part of Neo4j.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package neo4j

import (
	"fmt"
	"github.com/neo4j-drivers/gobolt"
)

type driverError struct {
	message string
}

func newDriverError(format string, args ...interface{}) error {
	return &driverError{message: fmt.Sprintf(format, args...)}
}

func isRetriableError(err error) bool {
	return gobolt.IsServiceUnavailable(err) || gobolt.IsTransientError(err) || gobolt.IsWriteError(err)
}

func (err *driverError) Error() string {
	return err.message
}

func IsServiceUnavailable(err error) bool {
	return gobolt.IsServiceUnavailable(err)
}

func IsDriverError(err error) bool {
	_, ok := err.(*driverError)
	return ok
}

func IsTransientError(err error) bool {
	return gobolt.IsTransientError(err)
}
