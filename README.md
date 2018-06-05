# CF Permissions BOSH Release

## Deploying perm with [cf-deployment](https://github.com/cloudfoundry/cf-deployment)

To deploy the Perm service use the following combination of opsfiles from [cf-deployment](https://github.com/cloudfoundry/cf-deployment)

NOTE: remove `bosh-lite.yml` and `disable-consul-bosh-lite.yml` to deploy to a real IaaS.

### Using `cf-mysql-release`
```
bosh -d cf deploy cf-deployment.yml \
  -v system_domain=$SYSTEM_DOMAIN \
  --vars-store deployment-vars.yml \
  -o cf-deployment/operations/experimental/enable-bpm.yml \
  -o cf-deployment/operations/experimental/skip-consul-cell-registrations.yml \
  -o cf-deployment/operations/experimental/skip-consul-locks.yml \
  -o cf-deployment/operations/experimental/use-bosh-dns.yml \
  -o cf-deployment/operations/experimental/disable-consul.yml \
  -o cf-deployment/operations/bosh-lite.yml \
  -o cf-deployment/operations/experimental/disable-consul-bosh-lite.yml \
  -o cf-deployment/operations/use-compiled-releases.yml \
  -o cf-deployment/operations/experimental/enable-mysql-tls.yml \
  -o cf-deployment/operations/experimental/perm-service.yml
```

### Using `pxc-release`
```
bosh -d cf deploy cf-deployment.yml \
  -v system_domain=$SYSTEM_DOMAIN \
  --vars-store deployment-vars.yml \
  -o cf-deployment/operations/experimental/enable-bpm.yml \
  -o cf-deployment/operations/experimental/skip-consul-cell-registrations.yml \
  -o cf-deployment/operations/experimental/skip-consul-locks.yml \
  -o cf-deployment/operations/experimental/use-bosh-dns.yml \
  -o cf-deployment/operations/experimental/disable-consul.yml \
  -o cf-deployment/operations/bosh-lite.yml \
  -o cf-deployment/operations/experimental/disable-consul-bosh-lite.yml \
  -o cf-deployment/operations/use-compiled-releases.yml \
  -o cf-deployment/operations/experimental/migrate-cf-mysql-to-pxc.yml \
  -o cf-deployment/operations/experimental/perm-service-with-pxc-release.yml \
  -o cf-deployment/operations/experimental/perm-service.yml
```
