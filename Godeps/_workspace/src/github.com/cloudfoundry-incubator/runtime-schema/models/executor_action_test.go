package models_test

import (
	"encoding/json"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry-incubator/runtime-schema/models"
)

var _ = Describe("ExecutorAction", func() {
	Describe("With an invalid action", func() {
		It("should fail to marshal", func() {
			invalidAction := []string{"butts", "from", "mars"}
			payload, err := json.Marshal(&ExecutorAction{Action: invalidAction})
			Ω(payload).Should(BeZero())
			Ω(err.(*json.MarshalerError).Err).Should(Equal(InvalidActionConversion))
		})

		It("should fail to unmarshal", func() {
			var unmarshalledAction *ExecutorAction
			actionPayload := `{"action":"buttz","args":{"from":"space"}}`
			err := json.Unmarshal([]byte(actionPayload), &unmarshalledAction)
			Ω(err).Should(Equal(InvalidActionConversion))
		})
	})

	itSerializesAndDeserializes := func(actionPayload string, action interface{}) {
		Describe("Converting to JSON", func() {
			It("creates a json representation of the object", func() {
				marshalledAction := action

				json, err := json.Marshal(&marshalledAction)
				Ω(err).Should(BeNil())
				Ω(json).Should(MatchJSON(actionPayload))
			})
		})

		Describe("Converting from JSON", func() {
			It("constructs an object from the json string", func() {
				var unmarshalledAction *ExecutorAction
				err := json.Unmarshal([]byte(actionPayload), &unmarshalledAction)
				Ω(err).Should(BeNil())
				Ω(*unmarshalledAction).Should(Equal(action))
			})
		})
	}

	Describe("Download", func() {
		itSerializesAndDeserializes(
			`{
				"action": "download",
				"args": {
					"from": "web_location",
					"to": "local_location",
					"extract": true
				}
			}`,
			ExecutorAction{
				Action: DownloadAction{
					From:    "web_location",
					To:      "local_location",
					Extract: true,
				},
			},
		)
	})

	Describe("Upload", func() {
		itSerializesAndDeserializes(
			`{
				"action": "upload",
				"args": {
					"from": "local_location",
					"to": "web_location"
				}
			}`,
			ExecutorAction{
				Action: UploadAction{
					From: "local_location",
					To:   "web_location",
				},
			},
		)
	})

	Describe("Run", func() {
		itSerializesAndDeserializes(
			`{
				"action": "run",
				"args": {
					"script": "rm -rf /",
					"timeout": 10000000,
					"env": [
						["FOO", "1"],
						["BAR", "2"]
					]
				}
			}`,
			ExecutorAction{
				Action: RunAction{
					Script:  "rm -rf /",
					Timeout: 10 * time.Millisecond,
					Env: [][]string{
						{"FOO", "1"},
						{"BAR", "2"},
					},
				},
			},
		)
	})

	Describe("FetchResult", func() {
		itSerializesAndDeserializes(
			`{
				"action": "fetch_result",
				"args": {
					"file": "/tmp/foo"
				}
			}`,
			ExecutorAction{
				Action: FetchResultAction{
					File: "/tmp/foo",
				},
			},
		)
	})
})
