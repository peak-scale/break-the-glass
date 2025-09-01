# Reference

Packages:

- [addons.projectcapsule.dev/v1alpha1](#addonsprojectcapsuledevv1alpha1)

# addons.projectcapsule.dev/v1alpha1

Resource Types:

- [BreakRequest](#breakrequest)

- [BreakRequestTemplate](#breakrequesttemplate)




## BreakRequest






BreakRequest is the Schema for the BreakRequests API.

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **apiVersion** | string | addons.projectcapsule.dev/v1alpha1 | true |
| **kind** | string | BreakRequest | true |
| **[metadata](https://kubernetes.io/docs/reference/generated/kubernetes-api/latest/#objectmeta-v1-meta)** | object | Refer to the Kubernetes API documentation for the fields of the `metadata` field. | true |
| **[spec](#breakrequestspec)** | object | BreakRequestSpec defines the desired state of BreakRequest. | false |
| **[status](#breakrequeststatus)** | object | BreakRequestStatus defines the observed state of BreakRequest. | false |


### BreakRequest.spec



BreakRequestSpec defines the desired state of BreakRequest.

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **templateName** | string | TemplateName the name of the template to use for this request | true |
| **duration** | string | The duration this BreakRequest should be valid for.
If no duration was defined the lifecycle is bound to the request itself -
if the request is deleted, it's the end of the duration.
The Request can also be Terminated by another automation via calling the ExpireRequest() API-Function. | false |
| **items** | map[string]object | Params the parameters to use for the template. | false |
| **reason** | string | A reason on why the request is needed | false |
| **[requestor](#breakrequestspecrequestor)** | object | Requesting actor for the access request. | false |
| **startTime** | string | Optional point in time when the access should begin. Must be in the future.
If omitted, this is set to the current time. The Request must already be approved before the start time.<br/><i>Format</i>: date-time<br/> | false |


### BreakRequest.spec.requestor



Requesting actor for the access request.

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **name** | string | The name of the entity | false |
| **type** | enum | The type of the entity<br/><i>Enum</i>: User, Group, System<br/> | false |


### BreakRequest.status



BreakRequestStatus defines the observed state of BreakRequest.

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **[active](#breakrequeststatusactive)** | object | Shows timestamps beetwen approval and termination of the request. | false |
| **[approved](#breakrequeststatusapproved)** | object | The Approved properties are set when the request is approved. | false |
| **[conditions](#breakrequeststatusconditionsindex)** | []object | conditions applied to the request.
Known conditions are "Requested", "Pending", "Denied", "Approved", "Active" and "Expired".
Latests condition is reflected in the phase. | false |
| **keepFor** | string | The duration this BreakRequest will be kept in the system after it has been expired (eg. auditing purposes)
If not set, the BreakRequest will be deleted after expiring. | false |
| **keepUntil** | string | The time when the request was created.<br/><i>Format</i>: date-time<br/> | false |
| **phase** | enum | <br/><i>Enum</i>: Requested, Pending, Denied, Approved, Active, Expired<br/> | false |
| **[review](#breakrequeststatusreview)** | object | Reviewer refers to the subject that either approved or denied the request | false |


### BreakRequest.status.active



Shows timestamps beetwen approval and termination of the request.

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **from** | string | <br/><i>Format</i>: date-time<br/> | false |
| **until** | string | <br/><i>Format</i>: date-time<br/> | false |


### BreakRequest.status.approved



The Approved properties are set when the request is approved.

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **duration** | string |  | false |
| **items** | map[string]object |  | false |
| **keepFor** | string |  | false |
| **startTime** | string | <br/><i>Format</i>: date-time<br/> | false |


### BreakRequest.status.conditions[index]



Condition contains details for one aspect of the current state of this API Resource.

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **lastTransitionTime** | string | lastTransitionTime is the last time the condition transitioned from one status to another.
This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable.<br/><i>Format</i>: date-time<br/> | true |
| **message** | string | message is a human readable message indicating details about the transition.
This may be an empty string. | true |
| **reason** | string | reason contains a programmatic identifier indicating the reason for the condition's last transition.
Producers of specific condition types may define expected values and meanings for this field,
and whether the values are considered a guaranteed API.
The value should be a CamelCase string.
This field may not be empty. | true |
| **status** | enum | status of the condition, one of True, False, Unknown.<br/><i>Enum</i>: True, False, Unknown<br/> | true |
| **type** | string | type of condition in CamelCase or in foo.example.com/CamelCase. | true |
| **observedGeneration** | integer | observedGeneration represents the .metadata.generation that the condition was set based upon.
For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date
with respect to the current state of the instance.<br/><i>Format</i>: int64<br/><i>Minimum</i>: 0<br/> | false |


### BreakRequest.status.review



Reviewer refers to the subject that either approved or denied the request

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **message** | string | Message with the review | false |
| **[reviewer](#breakrequeststatusreviewreviewer)** | object | The Entity revieweing this request | false |
| **verdict** | enum | The verdict made by the reviewing entity<br/><i>Enum</i>: Pending, Denied, Approved<br/> | false |


### BreakRequest.status.review.reviewer



The Entity revieweing this request

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **name** | string | The name of the entity | false |
| **type** | enum | The type of the entity<br/><i>Enum</i>: User, Group, System<br/> | false |

## BreakRequestTemplate






BreakRequestTemplate is the Schema for the breakrequesttemplates API.

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **apiVersion** | string | addons.projectcapsule.dev/v1alpha1 | true |
| **kind** | string | BreakRequestTemplate | true |
| **[metadata](https://kubernetes.io/docs/reference/generated/kubernetes-api/latest/#objectmeta-v1-meta)** | object | Refer to the Kubernetes API documentation for the fields of the `metadata` field. | true |
| **[spec](#breakrequesttemplatespec)** | object | BreakRequestTemplateSpec defines the desired state of BreakRequestTemplate. | false |
| **status** | object | BreakRequestTemplateStatus defines the observed state of BreakRequestTemplate. | false |


### BreakRequestTemplate.spec



BreakRequestTemplateSpec defines the desired state of BreakRequestTemplate.

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **[items](#breakrequesttemplatespecitemskey)** | map[string]object | Actual Items being created by this template | true |
| **approvalCondition** | string | ApprovalCondition an optional CEL expression that must be successful for the request to be approved. | false |
| **autoApprove** | boolean | AutoApprove requests created by this template will be automatically approved. | false |
| **duration** | string | The default duration of the BreakRequest referencing this template should be valid for. | false |
| **keepFor** | string | The duration of this AccessRequest will be kept in the system after it has been expired (eg. auditing purposes)
If not set, the AccessRequest will be deleted after expiring. | false |


### BreakRequestTemplate.spec.items[key]



TemplateItem

| **Name** | **Type** | **Description** | **Required** |
| :---- | :---- | :----------- | :-------- |
| **item** | object | Item | true |
| **paramSchema** | object | ParamSchema | false |
