# Manual Tests

## Install CRD

```bash
kubectl apply \
  -f config/crd/addons.projectcapsule.dev_breakrequests.yaml \
  -f config/crd/addons.projectcapsule.dev_breakrequesttemplates.yaml
```

## Apply manifests

```bash
kubectl apply -f testdata/manual
```