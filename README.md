# ISC config parser

It's an abstract to parser config files of ISC utilities (like `isc-dhcp-server` or `bind9`)

It should be used directly. It should be used to write on parsers of specific ISC utilities.

# Example

```
go get github.com/xaionaro-go/isccfg

# This's just an example utility that prints the config in JSON format
go install github.com/xaionaro-go/isccfg/iscparser

"$GOPATH"/bin/iscparser /etc/bind/named.conf.options
{
  "options": {
    "auth-nxdomain": {
      "_value": [
        "no"
      ]
    },
    "directory": {
      "_value": [
        "/var/cache/bind"
      ]
    },
    "dnssec-validation": {
      "_value": [
        "auto"
      ]
    },
    "listen-on-v6": {
      "_value": [
        "any"
      ]
    }
  }
}

$GOPATH/bin/iscparser /etc/dhcp/dhcpd.conf 
{
  "ddns-update-style": {
    "_value": [
      "none"
    ]
  },
  "default-lease-time": {
    "_value": [
      "600"
    ]
  },
  "max-lease-time": {
    "_value": [
      "7200"
    ]
  },
  "option": {
    "domain-name": {
      "_value": [
        "example.org"
      ]
    },
    "domain-name-servers": {
      "_value": [
        "ns1.example.org",
        "ns2.example.org"
      ]
    }
  },
  "subnet": {
    "10.5.5.0": {
      "netmask": {
        "255.255.255.224": {
          "default-lease-time": {
            "_value": [
              "600"
            ]
          },
          "max-lease-time": {
            "_value": [
              "7200"
            ]
          },
          "option": {
            "broadcast-address": {
              "_value": [
                "10.5.5.31"
              ]
            },
            "domain-name": {
              "_value": [
                "internal.example.org"
              ]
            },
            "domain-name-servers": {
              "_value": [
                "ns1.internal.example.org"
              ]
            },
            "routers": {
              "_value": [
                "10.5.5.1"
              ]
            }
          },
          "range": {
            "10.5.5.26": {
              "_value": [
                "10.5.5.30"
              ]
            }
          }
        }
      }
    }
  }
}
```

