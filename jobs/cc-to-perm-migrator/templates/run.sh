#!/bin/bash

set -e

/var/vcap/jobs/bpm/bin/bpm run cc-to-perm-migrator
