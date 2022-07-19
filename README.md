# karma

## Deploys

To remote exec:
```
gcloud compute ssh --zone "<ZONE>" "<VM NAME>"  --tunnel-through-iap --project "<PROJECTID>" --command="VERSION=something bash -s" < ./scripts/pull_and_deploy_gcp.sh
```
