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
	"github.com/neo4j-drivers/gobolt"
	"time"
)

// Result provides access to the result of the executing statement
type Result struct {
	keys            []string
	records         []Record
	current         *Record
	summary         ResultSummary
	runner          *statementRunner
	err             error
	runHandle       gobolt.RequestHandle
	runCompleted    bool
	resultHandle    gobolt.RequestHandle
	resultCompleted bool
}

func (result *Result) collectMetadata(metadata map[string]interface{}) {
	if metadata != nil {
		if resultAvailabilityTimer, ok := metadata["result_available_after"]; ok {
			result.summary.resultAvailableAfter = time.Duration(resultAvailabilityTimer.(int64)) / time.Millisecond
		}

		if resultConsumptionTimer, ok := metadata["result_consumed_after"]; ok {
			result.summary.resultConsumedAfter = time.Duration(resultConsumptionTimer.(int64)) / time.Millisecond
		}

		if typeString, ok := metadata["type"]; ok {
			switch typeString.(string) {
			case "r":
				result.summary.statementType = StatementTypeReadOnly
			case "rw":
				result.summary.statementType = StatementTypeReadWrite
			case "w":
				result.summary.statementType = StatementTypeWriteOnly
			case "s":
				result.summary.statementType = StatementTypeSchemaWrite
			default:
				result.summary.statementType = StatementTypeUnknown
			}
		}

		if stats, ok := metadata["stats"]; ok {
			if statsDict, ok := stats.(map[string]interface{}); ok {
				collectCounters(&statsDict, &result.summary.counters)
			}
		}

		if plan, ok := metadata["plan"]; ok {
			if plansDict, ok := plan.(map[string]interface{}); ok {
				result.summary.plan = collectPlan(&plansDict)
			}
		}

		if profile, ok := metadata["profile"]; ok {
			if profileDict, ok := profile.(map[string]interface{}); ok {
				result.summary.profile = collectProfile(&profileDict)
			}
		}

		if notifications, ok := metadata["notifications"]; ok {
			if notificationsList, ok := notifications.([]interface{}); ok {
				collectNotification(&notificationsList, &result.summary.notifications)
			}
		}
	}
}

func (result *Result) collectRecord(fields []interface{}) {
	if fields != nil {
		result.records = append(result.records, Record{keys: result.keys, values: fields})
	}
}

// Keys returns the keys available on the result set
func (result *Result) Keys() ([]string, error) {
	for !result.runCompleted {
		if currentResult, err := result.runner.receive(); currentResult == result && err != nil {
			return nil, err
		}
	}

	return result.keys, nil
}

// Next returns true only if there is a record to be processed
func (result *Result) Next() bool {
	if result.err != nil {
		return false
	}

	for !result.runCompleted {
		if currentResult, err := result.runner.receive(); currentResult == result && err != nil {
			return false
		}
	}

	for !result.resultCompleted && len(result.records) == 0 {
		if currentResult, err := result.runner.receive(); currentResult == result && err != nil {
			return false
		}
	}

	if len(result.records) > 0 {
		result.current = &result.records[0]
		result.records = result.records[1:]
	} else {
		result.current = nil
	}

	return result.current != nil
}

// Err returns the latest error that caused this Next to return false
func (result *Result) Err() error {
	return result.err
}

// Record returns the current record
func (result *Result) Record() *Record {
	return result.current
}

// Summary returns the summary information about the statement execution
func (result *Result) Summary() (*ResultSummary, error) {
	for result.err == nil && !result.resultCompleted {
		if _, err := result.runner.receive(); err != nil {
			result.err = err

			break
		}
	}

	return &result.summary, result.err
}

// Consume consumes the entire result and returns the summary information
// about the statement execution
func (result *Result) Consume() (*ResultSummary, error) {
	for result.Next() {

	}

	return &result.summary, result.err
}
