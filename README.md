# self-cli
A cli for managing and administering self applications


## Available commands

```sh
$ self-cli -h
```

## List all devices
To list all devices and their status:

```sh
$ self-cli device list
```

## Create a new device
To create a new device, you will need to run the following command and provide a valid device secret key. A new device key will be generated for you:
```sh
$ self-cli device create --secret-key MY-SECRET-KEY [appID]
```

To create a new device while providing your own device public key:
```sh
$ self-cli device create --secret-key MY-SECRET-KEY --device-public-key MY-NEW-DEVICE-PUBLIC-KEY [appID]
```

## Revoke an existing device
To revoke an existing device:
```sh
$ self-cli device revoke --secret-key MY-SECRET-KEY --effective-from 1607607355 [appID] [deviceID]
```

If your device key becomes compromised and you wish to retroactively revoke a device, you can specify a unix timestamp of when you want the revocation to take place:
```sh
$ self-cli device revoke --secret-key MY-SECRET-KEY --effective-from 1607607355 [appID] [deviceID]
```

## Rotate a devices keys
To rotate the keys on an existing device, you can run the following:
```sh
$ self-cli device rotate --secret-key MY-SECRET-KEY [appID] [deviceID]
```

If you wish to provide the public key for the device yourself, you can run:
```sh
$ self-cli device rotate --secret-key MY-SECRET-KEY --device-public-key MY-NEW-DEVICE-PUBLIC-KEY [appID] [deviceID]
```

## Account recovery
If you have lost access to your account and wish to recover your account, you can use the following command. It will revoke all existing keys for your account and create you a new device and recovery keypair:
```sh
$ self-cli identity recover --recovery-key MY-RECOVERY-KEY [appID]
```