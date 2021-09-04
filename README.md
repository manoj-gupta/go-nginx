# Readme

## Certificate

Our proxy will be an HTTPS server (--proto https), so we need certificate and private key. Let’s use self-signed certificate.

## Generate self-signed certificate

* cd certs
* ./gen.sh

## Add self-signed certificate as trusted to OS

### OSX
* https://tosbourn.com/getting-os-x-to-trust-self-signed-ssl-certificates/

### Linux
Copy your certificate in PEM format (the format that has ```----BEGIN CERTIFICATE----``` in it) into ```/usr/local/share/ca-certificates``` and name it with a ```.crt``` file extension.

Then run ```sudo update-ca-certificates```.

## Run Server
* cd server
* export variables 
    * AUTH_USERNAME/AUTH_PASSWORD (client/server)
    * ARTIFACTORY_USERNAME/ARTIFACTORY_PASSWORD (artifactory credentials - optional)
* go run main.go

## Run Client
* cd client
* export variables
    * AUTH_USERNAME/AUTH_PASSWORD (client/server)
* go run client.go

# NGINX

You can use **nginx** in place of server to test various mode of authentication.

## Components

* ```assets``` - this folder contains the static files to be served for different handlers
    * config - serves static *index.html* files for unprotected *config* handler
    * test - serves static *index.html* files for unprotected *test* handler
    * basic_auth - serves static *index.html* files for userid/password protected *basic_auth* handler
    * jwt_auth - serves static *index.html* files for jwt protected *jwt_auth* handlers

