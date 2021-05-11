// Copyright 2020 IBM Corp.
// SPDX-License-Identifier: Apache-2.0

package taxonomy

import (
	"testing"
)

var (
	PMTaxStructsName = "policymanager.structs.schema.json"
	PMTaxValsName    = "policymanager.values.schema.json"

	intentGood = "{\"intent\":\"Marketing\"}"
	intentBad  = "{\"intent\":\"whatever\"}"

	roleGood = "{\"role\":\"Data Scientist\"}"
	roleBad  = "{\"role\":\"whatever\"}"

	// {"request_context":{"intent":"Marketing", "role":"Data Scientist"}}
	requestContextGood = "{\"request_context\":{\"intent\":\"Marketing\", \"role\":\"Data Scientist\"}}"

	// {"request_context":{"intent":"Marketing", "role":"Data Scientist", "x":"Y"}}
	requestContextGoodExtraProps = "{\"request_context\":{\"intent\":\"Marketing\", \"role\":\"Data Scientist\", \"x\":\"Y\"}}"

	// {"request_context":{"role":"Data Scientist", "x":"Y"}}
	requestContextBadNoIntent = "{\"request_context\":{\"role\":\"Data Scientist\", \"x\":\"Y\"}}"

	// {"action_type":"read"}
	actionTypeGood = "{\"action_type\":\"read\"}"

	// {"action_type":"xxx"}
	actionTypeBad = "{\"action_type\":\"xxx\"}"

	// {"action":{"action_type":"read", "processingLocation":"Turkey"}}
	actionGood = "{\"action\":{\"action_type\":\"read\", \"processingLocation\":\"Turkey\"}}"

	// {"action":{"action_type":"read", "processingLocation":"xx"}}
	actionBad = "{\"action\":{\"action_type\":\"read\", \"processingLocation\":\"xx\"}}"

	// {"governance_decision_request": {"request_context":{"intent":"Marketing", "role":"Data Scientist"},"action":{"action_type":"read", "processingLocation":"Turkey"}, "resource":{"name":"file1"}}}
	governanceRequestGood = "{\"governance_decision_request\": {\"request_context\":{\"intent\":\"Marketing\", \"role\":\"Data Scientist\"},\"action\":{\"action_type\":\"read\", \"processingLocation\":\"Turkey\"}, \"resource\":{\"name\":\"file1\"}}}"

	// {"governance_decision_request": {"request_context":{"intent":"Marketing", "role":"Data Scientist"},"action":{"action_type":"read", "processingLocation":"Turkey"}}}
	governanceRequestBadNoResource = "{\"governance_decision_request\": {\"request_context\":{\"intent\":\"Marketing\", \"role\":\"Data Scientist\"},\"action\":{\"action_type\":\"read\", \"processingLocation\":\"Turkey\"}}}"

	// {\"governance_decision_response\": {\"decision_id\":\"abcde1234\", \"governance_actions\":[{\"actions\":\"\", \"used_policy\":\"policyID112233\"}]}}
	governanceResponseBadNoDecision = "{\"governance_decision_response\": {\"decision_id\":\"abcde1234\", \"governance_actions\":[{\"actions\": {\"name\":\"\"}, \"used_policy\":\"policyID112233\"}]}}"

	// {"governance_decision_response": {"resource":{"name":"file1"}, "governance_decision":"allow"}}
	governanceResponseGood = "{\"governance_decision_response\": {\"decision_id\":\"abcde1234\", \"governance_actions\":[{\"actions\": {\"name\":\"RedactColumn\", \"params\":[\"Column 1\"]}, \"used_policy\":\"policyID112233\"}]}}"
)

func TestPolicyManagerTaxonomy(t *testing.T) {
	ValidateTaxonomy(t, PMTaxValsName, intentGood, "intentGood", true)
	ValidateTaxonomy(t, PMTaxValsName, intentBad, "intentBad", false)
	ValidateTaxonomy(t, PMTaxValsName, roleGood, "roleGood", true)
	ValidateTaxonomy(t, PMTaxValsName, roleBad, "roleBad", false)
	ValidateTaxonomy(t, PMTaxValsName, requestContextGood, "requestContextGood", true)
	ValidateTaxonomy(t, PMTaxValsName, requestContextGoodExtraProps, "requestContextGoodExtraProps", true)
	ValidateTaxonomy(t, PMTaxValsName, requestContextBadNoIntent, "requestContextBadNoIntent", false)
	ValidateTaxonomy(t, PMTaxValsName, actionTypeGood, "actionTypeGood", true)
	ValidateTaxonomy(t, PMTaxValsName, actionTypeBad, "actionTypeBad", false)

	ValidateTaxonomy(t, PMTaxStructsName, actionGood, "actionGood", true)
	ValidateTaxonomy(t, PMTaxStructsName, actionBad, "actionBad", false)
	ValidateTaxonomy(t, PMTaxStructsName, governanceRequestGood, "governanceRequestGood", true)
	ValidateTaxonomy(t, PMTaxStructsName, governanceRequestBadNoResource, "governanceRequestBadNoResource", false)
	ValidateTaxonomy(t, PMTaxStructsName, governanceResponseBadNoDecision, "governanceResponseBadNoDecision", false)
	ValidateTaxonomy(t, PMTaxStructsName, governanceResponseGood, "governanceResponseGood", true)
}
