package mysql

// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Code generated by Microsoft (R) AutoRest Code Generator.
// Changes may cause incorrect behavior and will be lost if the code is regenerated.

import (
	"context"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/validation"
	"net/http"
)

// CheckNameAvailabilityClient is the the Microsoft Azure management API provides create, read, update, and delete
// functionality for Azure MySQL resources including servers, databases, firewall rules, VNET rules, log files and
// configurations.
type CheckNameAvailabilityClient struct {
	BaseClient
}

// NewCheckNameAvailabilityClient creates an instance of the CheckNameAvailabilityClient client.
func NewCheckNameAvailabilityClient(subscriptionID string) CheckNameAvailabilityClient {
	return NewCheckNameAvailabilityClientWithBaseURI(DefaultBaseURI, subscriptionID)
}

// NewCheckNameAvailabilityClientWithBaseURI creates an instance of the CheckNameAvailabilityClient client.
func NewCheckNameAvailabilityClientWithBaseURI(baseURI string, subscriptionID string) CheckNameAvailabilityClient {
	return CheckNameAvailabilityClient{NewWithBaseURI(baseURI, subscriptionID)}
}

// Execute check the availability of name for resource
//
// nameAvailabilityRequest is the required parameters for checking if resource name is available.
func (client CheckNameAvailabilityClient) Execute(ctx context.Context, nameAvailabilityRequest NameAvailabilityRequest) (result NameAvailability, err error) {
	if err := validation.Validate([]validation.Validation{
		{TargetValue: nameAvailabilityRequest,
			Constraints: []validation.Constraint{{Target: "nameAvailabilityRequest.Name", Name: validation.Null, Rule: true, Chain: nil}}}}); err != nil {
		return result, validation.NewErrorWithValidationError(err, "mysql.CheckNameAvailabilityClient", "Execute")
	}

	req, err := client.ExecutePreparer(ctx, nameAvailabilityRequest)
	if err != nil {
		err = autorest.NewErrorWithError(err, "mysql.CheckNameAvailabilityClient", "Execute", nil, "Failure preparing request")
		return
	}

	resp, err := client.ExecuteSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		err = autorest.NewErrorWithError(err, "mysql.CheckNameAvailabilityClient", "Execute", resp, "Failure sending request")
		return
	}

	result, err = client.ExecuteResponder(resp)
	if err != nil {
		err = autorest.NewErrorWithError(err, "mysql.CheckNameAvailabilityClient", "Execute", resp, "Failure responding to request")
	}

	return
}

// ExecutePreparer prepares the Execute request.
func (client CheckNameAvailabilityClient) ExecutePreparer(ctx context.Context, nameAvailabilityRequest NameAvailabilityRequest) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"subscriptionId": autorest.Encode("path", client.SubscriptionID),
	}

	const APIVersion = "2017-04-30-preview"
	queryParameters := map[string]interface{}{
		"api-version": APIVersion,
	}

	preparer := autorest.CreatePreparer(
		autorest.AsJSON(),
		autorest.AsPost(),
		autorest.WithBaseURL(client.BaseURI),
		autorest.WithPathParameters("/subscriptions/{subscriptionId}/providers/Microsoft.DBforMySQL/checkNameAvailability", pathParameters),
		autorest.WithJSON(nameAvailabilityRequest),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare((&http.Request{}).WithContext(ctx))
}

// ExecuteSender sends the Execute request. The method will close the
// http.Response Body if it receives an error.
func (client CheckNameAvailabilityClient) ExecuteSender(req *http.Request) (*http.Response, error) {
	return autorest.SendWithSender(client, req,
		azure.DoRetryWithRegistration(client.Client))
}

// ExecuteResponder handles the response to the Execute request. The method always
// closes the http.Response Body.
func (client CheckNameAvailabilityClient) ExecuteResponder(resp *http.Response) (result NameAvailability, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result),
		autorest.ByClosing())
	result.Response = autorest.Response{Response: resp}
	return
}
