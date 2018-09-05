#!/bin/bash

set -e

/var/vcap/jobs/bpm/bin/bpm run perm-migrate-down
/var/vcap/jobs/bpm/bin/bpm run perm-migrate-up
