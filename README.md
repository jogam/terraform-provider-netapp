Terraform Provider
==================

- Under Construction
- Current Status: key/value storage from GO to python with proper start/stop etc

<img src="https://cdn.rawgit.com/hashicorp/terraform-website/master/content/source/assets/images/logo-hashicorp.svg" width="600px">

Maintainers
-----------

This provider plugin is being developped by jogam.

Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) 0.11.9+ (0.11.9 used for dev/testing)
-	[Go](https://golang.org/doc/install) 1.10.? (1.8 to build the provider plugin, 1.10.1 used for dev/testing)

Usage
---------------------

```
TBA
```

Building The Provider
---------------------

Clone repository to: `$GOPATH/src/github.com/terraform-providers/terraform-provider-template`

```sh
$ mkdir -p $GOPATH/src/github.com/terraform-providers; cd $GOPATH/src/github.com/terraform-providers
$ git clone git@github.com:terraform-providers/terraform-provider-template
```

Enter the provider directory and build the provider

```sh
$ cd $GOPATH/src/github.com/terraform-providers/terraform-provider-template
$ make build
```

Using the provider
----------------------
## Fill in for each provider
TBA, 
- need python/virtualenv
- supported NetApp device for access to NetApp Manageability SDK download

Developing the Provider
---------------------------

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (version 1.8+ is *required*). You'll also need to correctly setup a [GOPATH](http://golang.org/doc/code.html#GOPATH), as well as adding `$GOPATH/bin` to your `$PATH`.

To compile the provider, run `make build`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

```sh
$ make bin
...
$ $GOPATH/bin/terraform-provider-template
...
```

In order to test the provider, you can simply run `make test`.

```sh
$ make test
```

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```sh
$ make testacc
```
