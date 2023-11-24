# Instructions

```bash
# Build the docker image
$ docker build -t helloserver .

# Run the docker image
$ docker run -d --rm --name kauppie-helloserver1 helloserver

# Check its IP address
$ docker exec kauppie-helloserver1 ifconfig eth0
```

Validate that [`inventory.ini`](inventory.ini) has the correct IP address.

I'm using key based method for SSH. Files `key` and `key.pub` contain the private and public keys respectively. `key.pub` is included in the docker image during build time. `key` is used by Ansible to connect to the server. Acknowledgement: keys should not be included in a git repository, but they are here just for testing purposes.

Use the following command to run the playbook:

```bash
$ ansible-playbook -i inventory.ini -K --extra-vars 'ansible_user=ssluser ansible_ssh_private_key_file=key' playbook.yaml
```

Ansible will ask for the password of the user `ssluser`. The password is `eee`. This will be later used to install `git` to the container. You will also need enter `yes` when asked to validate the server's fingerprint.

Adding the second server is done with above instructions, with the addition of the second container's IP address to the `inventory.ini` file on a new line. Here is an example of the inventory file I used for two servers:

```ini
[myhosts]
172.17.0.2
172.17.0.3
```

## Playbook outputs

### O1

```
PLAY [all] *******************************************************************************************************************************************************************

TASK [Gathering Facts] *******************************************************************************************************************************************************
ok: [172.17.0.2]

TASK [Ensure latest version of git is installed] *****************************************************************************************************************************
changed: [172.17.0.2]

TASK [Query uptime] **********************************************************************************************************************************************************
changed: [172.17.0.2]

PLAY RECAP *******************************************************************************************************************************************************************
172.17.0.2                 : ok=3    changed=2    unreachable=0    failed=0    skipped=0    rescued=0    ignored=0   
```

### O2

```
PLAY [all] *******************************************************************************************************************************************************************

TASK [Gathering Facts] *******************************************************************************************************************************************************
ok: [172.17.0.2]

TASK [Ensure latest version of git is installed] *****************************************************************************************************************************
ok: [172.17.0.2]

TASK [Query uptime] **********************************************************************************************************************************************************
changed: [172.17.0.2]

PLAY RECAP *******************************************************************************************************************************************************************
172.17.0.2                 : ok=3    changed=1    unreachable=0    failed=0    skipped=0    rescued=0    ignored=0   
```

### O3

```
PLAY [all] *******************************************************************************************************************************************************************

TASK [Gathering Facts] *******************************************************************************************************************************************************
ok: [172.17.0.2]
ok: [172.17.0.3]

TASK [Ensure latest version of git is installed] *****************************************************************************************************************************
ok: [172.17.0.2]
changed: [172.17.0.3]

TASK [Query uptime] **********************************************************************************************************************************************************
changed: [172.17.0.2]
changed: [172.17.0.3]

PLAY RECAP *******************************************************************************************************************************************************************
172.17.0.2                 : ok=3    changed=1    unreachable=0    failed=0    skipped=0    rescued=0    ignored=0   
172.17.0.3                 : ok=3    changed=2    unreachable=0    failed=0    skipped=0    rescued=0    ignored=0   
```

### O4

```
PLAY [all] *******************************************************************************************************************************************************************

TASK [Gathering Facts] *******************************************************************************************************************************************************
ok: [172.17.0.2]
ok: [172.17.0.3]

TASK [Ensure latest version of git is installed] *****************************************************************************************************************************
ok: [172.17.0.2]
ok: [172.17.0.3]

TASK [Query uptime] **********************************************************************************************************************************************************
changed: [172.17.0.2]
changed: [172.17.0.3]

PLAY RECAP *******************************************************************************************************************************************************************
172.17.0.2                 : ok=3    changed=1    unreachable=0    failed=0    skipped=0    rescued=0    ignored=0   
172.17.0.3                 : ok=3    changed=1    unreachable=0    failed=0    skipped=0    rescued=0    ignored=0   
```

## Analysis of `uptime`

One might expect to see the uptime to be printed to host machine's terminal. This is not the case since the commands are run on the remote machine, and only the status of the command is printed to the terminal. Since the output is different for each run of `uptime`, task `Query uptime` is marked as changed every time.

## What was easy and difficult

Adding key based authentication was easier than I thought, since it really only required making the key pair, putting the public key in the docker image and using the private key in the playbook.
Most difficult part was trying to figure out how to execute the playbook without needing the password for the user. Eventually I understood that this method of 'passwordlessness' was asked only for SSH,
and not for the commands run inside the container.
