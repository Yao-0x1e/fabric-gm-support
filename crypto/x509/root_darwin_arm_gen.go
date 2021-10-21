// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

// Generates root_darwin_armx.go.
//
// As of iOS 8, there is no API for querying the system trusted X.509 root
// certificates. We could use SecTrustEvaluate to verify that a trust chain
// exists for a certificate, but the x509 API requires returning the entire
// chain.
//
// Apple publishes the list of trusted root certificates for iOS on
// support.apple.com. So we parse the list and extract the certificates from
// an OS X machine and embed them into the x509 package.
package main

import (
	"bytes"
	"encoding/hex"
	"encoding/pem"
	"flag"
	"fmt"
	"github.com/gxnublockchain/gmsupport/crypto/sha256"
	"github.com/gxnublockchain/gmsupport/crypto/x509"
	"go/format"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
)

var output = flag.String("output", "root_darwin_armx.go", "file name to write")

func main() {
	certs, err := selectCerts()
	if err != nil {
		log.Fatal(err)
	}

	buf := new(bytes.Buffer)

	fmt.Fprintf(buf, "// Code generated by root_darwin_arm_gen --output %s; DO NOT EDIT.\n", *output)
	fmt.Fprintf(buf, "%s", header)

	fmt.Fprintf(buf, "const systemRootsPEM = `\n")
	for _, cert := range certs {
		b := &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: cert.Raw,
		}
		if err := pem.Encode(buf, b); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Fprintf(buf, "`")

	source, err := format.Source(buf.Bytes())
	if err != nil {
		log.Fatal("source format error:", err)
	}
	if err := ioutil.WriteFile(*output, source, 0644); err != nil {
		log.Fatal(err)
	}
}

func selectCerts() ([]*x509.Certificate, error) {
	ids, err := fetchCertIDs()
	if err != nil {
		return nil, err
	}

	scerts, err := sysCerts()
	if err != nil {
		return nil, err
	}

	var certs []*x509.Certificate
	for _, id := range ids {
		if c, ok := scerts[id.fingerprint]; ok {
			certs = append(certs, c)
		} else {
			fmt.Printf("WARNING: cannot find certificate: %s (fingerprint: %s)\n", id.name, id.fingerprint)
		}
	}
	return certs, nil
}

func sysCerts() (certs map[string]*x509.Certificate, err error) {
	cmd := exec.Command("/usr/bin/security", "find-certificate", "-a", "-p", "/System/Library/Keychains/SystemRootCertificates.keychain")
	data, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	certs = make(map[string]*x509.Certificate)
	for len(data) > 0 {
		var block *pem.Block
		block, data = pem.Decode(data)
		if block == nil {
			break
		}
		if block.Type != "CERTIFICATE" || len(block.Headers) != 0 {
			continue
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			continue
		}

		fingerprint := sha256.Sum256(cert.Raw)
		certs[hex.EncodeToString(fingerprint[:])] = cert
	}
	return certs, nil
}

type certID struct {
	name        string
	fingerprint string
}

// fetchCertIDs fetches IDs of iOS X509 certificates from apple.com.
func fetchCertIDs() ([]certID, error) {
	// Download the iOS 11 support page. The index for all iOS versions is here:
	// https://support.apple.com/en-us/HT204132
	resp, err := http.Get("https://support.apple.com/en-us/HT208125")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	text := string(body)
	text = text[strings.Index(text, "<div id=trusted"):]
	text = text[:strings.Index(text, "</div>")]

	var ids []certID
	cols := make(map[string]int)
	for i, rowmatch := range regexp.MustCompile("(?s)<tr>(.*?)</tr>").FindAllStringSubmatch(text, -1) {
		row := rowmatch[1]
		if i == 0 {
			// Parse table header row to extract column names
			for i, match := range regexp.MustCompile("(?s)<th>(.*?)</th>").FindAllStringSubmatch(row, -1) {
				cols[match[1]] = i
			}
			continue
		}

		values := regexp.MustCompile("(?s)<td>(.*?)</td>").FindAllStringSubmatch(row, -1)
		name := values[cols["Certificate name"]][1]
		fingerprint := values[cols["Fingerprint (SHA-256)"]][1]
		fingerprint = strings.ReplaceAll(fingerprint, "<br>", "")
		fingerprint = strings.ReplaceAll(fingerprint, "\n", "")
		fingerprint = strings.ReplaceAll(fingerprint, " ", "")
		fingerprint = strings.ToLower(fingerprint)

		ids = append(ids, certID{
			name:        name,
			fingerprint: fingerprint,
		})
	}
	return ids, nil
}

const header = `
// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build cgo
// +build darwin
// +build arm arm64 ios

package x509

func loadSystemRoots() (*CertPool, error) {
	p := NewCertPool()
	p.AppendCertsFromPEM([]byte(systemRootsPEM))
	return p, nil
}
`
