# Break-The-Glass

This guide provides instructions on how to use the Break-The-Glass controller to manage temporary, elevated privileges
in your Kubernetes cluster.

## Introduction

The Break-The-Glass controller provides a mechanism for granting temporary access to Kubernetes resources. This is
useful in emergency situations where a user or a system needs elevated permissions to perform a critical task. This
process is often referred to as "breaking the glass".

The controller introduces two custom resources:

* `BreakRequestTemplate`: A cluster-scoped resource that defines a reusable template for temporary access.
* `BreakRequest`: A namespaced resource that represents a request for temporary access based on a
  `BreakRequestTemplate`.

## Concepts

### BreakRequestTemplate

A `BreakRequestTemplate` is a blueprint for a type of temporary access. It defines:

* **What** resources will be created (e.g., a `RoleBinding`).
* **How long** the access will be granted (`defaultDuration`, `maxDuration`).
* **Whether** the request is automatically approved (`autoApprove`).
* **Optional conditions** for automatic approval (`approvalCondition`).
* **How long** to keep the `BreakRequest` for auditing purposes after it expires (`keepFor`).

`BreakRequestTemplate` is a cluster-scoped resource, meaning it can be used by `BreakRequest`s from any namespace.

### BreakRequest

A `BreakRequest` is a user's request for temporary access. It is created in a specific namespace and must reference a
`BreakRequestTemplate`. When a `BreakRequest` is created, it goes through the following lifecycle:

1. **Requested**: The initial state of the request.
2. **Pending**: The request is awaiting review (if not auto-approved).
3. **Approved**: The request has been approved by a reviewer.
4. **Active**: The temporary access is granted, and the resources defined in the template are created.
5. **Expired**: The access duration has passed, and the temporary resources are deleted.
6. **Denied**: The request has been denied by a reviewer.

## Using `BreakRequestTemplate`

Here is an example of a `BreakRequestTemplate`:

```yaml
apiVersion: addons.projectcapsule.dev/v1alpha1
kind: BreakRequestTemplate
metadata:
  name: cluster-admin-access
spec:
  defaultDuration: 1h
  maxDuration: 4h
  keepFor: 72h
  autoApprove: false
  items:
    rolebinding:
      manifestTemplate:
        apiVersion: rbac.authorization.k8s.io/v1
        kind: RoleBinding
        metadata:
          name: break-glass-{{ .request.metadata.name }}
          namespace: "{{ .request.metadata.namespace }}"
        roleRef:
          apiGroup: rbac.authorization.k8s.io
          kind: ClusterRole
          name: cluster-admin
        subjects:
          - kind: "{{ .request.spec.requestor.kind }}"
            name: "{{ .request.spec.requestor.name }}"
            apiGroup: rbac.authorization.k8s.io
```

### `BreakRequestTemplate` Specification

* `spec.defaultDuration` (required): The default duration for which the access is granted. This is used if the
  `BreakRequest` does not specify a duration.
* `spec.maxDuration`: The maximum duration for which the access can be granted. A `BreakRequest` cannot request a
  duration longer than this.
* `spec.keepFor`: How long to keep the `BreakRequest` resource after it has expired. This is useful for auditing
  purposes. If not set, the `BreakRequest` is deleted immediately after expiration.
* `spec.autoApprove`: If set to `true`, `BreakRequest`s using this template are automatically approved.
* `spec.approvalCondition`: A [CEL expression](https://kubernetes.io/docs/reference/using-api/cel/) that is evaluated to
  automatically approve a `BreakRequest`. The request is approved if the expression evaluates to `true`.
* `spec.items` (required): A map of items to be created when the `BreakRequest` is active. The keys of the map are the
  names of the items, and the values are the item definitions.
    * `manifestTemplate`: A Go template for the Kubernetes resource to be created. The template has access to the
      `BreakRequest` object via the `.request` variable.

## Using `BreakRequest`

Here is an example of a `BreakRequest` that uses the `cluster-admin-access` template:

```yaml
apiVersion: addons.projectcapsule.dev/v1alpha1
kind: BreakRequest
metadata:
  name: my-cluster-admin-request
  namespace: default
spec:
  templateName: cluster-admin-access
  reason: "Need to debug a critical issue in the cluster."
  requestor:
    kind: User
    name: "john.doe@example.com"
  duration: 30m
```

### `BreakRequest` Specification

* `spec.templateName` (required): The name of the `BreakRequestTemplate` to use.
* `spec.reason`: A human-readable reason for the access request.
* `spec.requestor`: The user or system requesting the access.
    * `kind`: The kind of the requestor (`User`, `Group`, or `ServiceAccount`).
    * `name`: The name of the requestor.
* `spec.duration`: The requested duration for the access. This cannot exceed the `maxDuration` specified in the
  template. If not specified, the `defaultDuration` from the template is used.
* `spec.startTime`: An optional timestamp for when the access should begin. If omitted, the access begins as soon as the
  request is approved.

### `BreakRequest` Status

The `status` of a `BreakRequest` provides information about its current state:

* `status.phase`: The current phase of the request (`Requested`, `Pending`, `Denied`, `Approved`, `Active`, `Expired`).
* `status.review`: Information about the review, including the verdict, the reviewer, and a message.
* `status.active`: The time period when the access is active.
* `status.conditions`: A list of conditions that the `BreakRequest` has gone through.
* `status.template`: The properties copied from the `BreakRequestTemplate`.
* `status.approved`: The properties set when the request is approved.
* `status.keepUntil`: The time until which the `BreakRequest` will be kept after expiration.

## CLI Plugin

The Break-The-Glass controller comes with a `kubectl` plugin to help manage `BreakRequest`s. The plugin provides the
following commands:

### `review`

The `review` command allows you to approve or deny a `BreakRequest`. You can use it in interactive or non-interactive
mode.

**Interactive Mode:**

```bash
break-the-glass review <break-request-name> -n <namespace>
```

This will display the details of the `BreakRequest` and prompt you to approve or deny it.

**Non-Interactive Mode:**

To approve a request:

```bash
break-the-glass review <break-request-name> -n <namespace> --approve
```

To deny a request:

```bash
break-the-glass review <break-request-name> -n <namespace> --deny
```

**Options:**

* `--approve`: Approve the request.
* `--deny`: Deny the request.
* `-m, --message`: Add a message to the review.
* `--start-time`: Set the start time for the access (e.g., `2025-07-15T14:45:00Z`).
* `--duration`: Override the duration of the access (e.g., `1h30m`).
* `--keep-for`: Override the duration for which the request is kept after expiration (e.g., `72h`).

### `activate`

The `activate` command manually activates an approved `BreakRequest`.

```bash
break-the-glass activate <break-request-name> -n <namespace>
```

### `expire`

The `expire` command manually expires an active `BreakRequest`.

```bash
break-the-glass expire <break-request-name> -n <namespace>
```
