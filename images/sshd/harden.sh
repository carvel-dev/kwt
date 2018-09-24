#!/usr/bin/env bash

# Based on https://github.com/cloudfoundry/bosh-linux-stemcell-builder/blob/741f485675f13ec3cda19d375b8f30b1aa1c584c/stemcell_builder/stages/base_ssh/apply.sh

set -e

chmod 0600 /etc/ssh/sshd_config

# protect against as-shipped sshd_config that has no newline at end
echo "" >> /etc/ssh/sshd_config

sed "/^ *UseDNS/d" -i /etc/ssh/sshd_config
echo 'UseDNS no' >> /etc/ssh/sshd_config

sed "/^ *PermitRootLogin/d" -i /etc/ssh/sshd_config
echo 'PermitRootLogin no' >> /etc/ssh/sshd_config

sed "/^ *X11Forwarding/d" -i /etc/ssh/sshd_config
sed "/^ *X11DisplayOffset/d" -i /etc/ssh/sshd_config
echo 'X11Forwarding no' >> /etc/ssh/sshd_config

sed "/^ *MaxAuthTries/d" -i /etc/ssh/sshd_config
echo 'MaxAuthTries 3' >> /etc/ssh/sshd_config

sed "/^ *PermitEmptyPasswords/d" -i /etc/ssh/sshd_config
echo 'PermitEmptyPasswords no' >> /etc/ssh/sshd_config

sed "/^ *Protocol/d" -i /etc/ssh/sshd_config
echo 'Protocol 2' >> /etc/ssh/sshd_config

sed "/^ *HostbasedAuthentication/d" -i /etc/ssh/sshd_config
echo 'HostbasedAuthentication no' >> /etc/ssh/sshd_config

sed "/^ *Banner/d" -i /etc/ssh/sshd_config
echo 'Banner /etc/issue.net' >> /etc/ssh/sshd_config

sed "/^ *IgnoreRhosts/d" -i /etc/ssh/sshd_config
echo 'IgnoreRhosts yes' >> /etc/ssh/sshd_config

sed "/^ *ClientAliveInterval/d" -i /etc/ssh/sshd_config
echo 'ClientAliveInterval 900' >> /etc/ssh/sshd_config

sed "/^ *PermitUserEnvironment/d" -i /etc/ssh/sshd_config
echo 'PermitUserEnvironment no' >> /etc/ssh/sshd_config

sed "/^ *ClientAliveCountMax/d" -i /etc/ssh/sshd_config
echo 'ClientAliveCountMax 0' >> /etc/ssh/sshd_config

sed "/^ *PasswordAuthentication/d" -i /etc/ssh/sshd_config
echo 'PasswordAuthentication no' >> /etc/ssh/sshd_config

sed "/^ *PrintLastLog/d" -i /etc/ssh/sshd_config
echo 'PrintLastLog yes' >> /etc/ssh/sshd_config

sed "/^ *DenyUsers/d" -i /etc/ssh/sshd_config
echo 'DenyUsers root' >> /etc/ssh/sshd_config

sed "/^[ #]*HostKey \/etc\/ssh\/ssh_host_dsa_key/d" -i /etc/ssh/sshd_config
for type in {rsa,ecdsa,ed25519}; do
  sed "s/^[ #]*HostKey \/etc\/ssh\/ssh_host_${type}_key/HostKey \/etc\/ssh\/ssh_host_${type}_key/" -i /etc/ssh/sshd_config
done

#  Allow only 3DES and AES series ciphers
sed "/^ *Ciphers/d" -i /etc/ssh/sshd_config
echo 'Ciphers aes256-gcm@openssh.com,aes128-gcm@openssh.com,aes256-ctr,aes192-ctr,aes128-ctr' >> /etc/ssh/sshd_config

# Disallow Weak MACs
sed "/^ *MACs/d" -i /etc/ssh/sshd_config
echo 'MACs hmac-sha2-512-etm@openssh.com,hmac-sha2-256-etm@openssh.com,hmac-ripemd160-etm@openssh.com,umac-128-etm@openssh.com,hmac-sha2-512,hmac-sha2-256,hmac-ripemd160,hmac-sha1' >> /etc/ssh/sshd_config

cat << EOF > /etc/issue
Unauthorized use is strictly prohibited. All access and activity
is subject to logging and monitoring.
EOF

cp /etc/issue{,.net}
