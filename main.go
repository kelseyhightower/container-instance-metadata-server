package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
)

var (
	defaultCredentials  string
	metadataFile        string
	serviceAccountEmail string
	listenAddress       string
)

func main() {
	flag.StringVar(&defaultCredentials, "google-application-credentials", "", "Default service account file path")
	flag.StringVar(&listenAddress, "listen-address", "127.0.0.1:8888", "The HTTP listen address")
	flag.StringVar(&metadataFile, "metadata", "metadata.json", "Metadata file path")
	flag.StringVar(&serviceAccountEmail, "service-account", "", "The email address of an IAM service account")
	flag.Parse()

	log.Println("Starting Container Instance Metadata Service ...")
	log.Printf("Impersonating %s", serviceAccountEmail)
	log.Printf("Listening on %s", listenAddress)

	md, err := MetadataFromFile(metadataFile)
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/", &metadataHandler{md, serviceAccountEmail})

	log.Fatal(http.ListenAndServe(listenAddress, nil))
}

type metadataHandler struct {
	md *Metadata
	sa string
}

func (h *metadataHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling request for %s", r.URL.RequestURI())

	w.Header().Set("Metadata-Flavor", "Google")
	w.Header().Set("Server", "Metadata Server for Serverless")
	w.Header().Set("Content-Type", "application/text")

	if r.Header.Get("Metadata-Flavor") != "Google" {
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")
		w.WriteHeader(403)
		return
	}

	switch r.URL.Path {
	case "/computeMetadata/v1/instance/id":
		fmt.Fprintf(w, h.md.Instance.ID)
	case "/computeMetadata/v1/instance/service-accounts/default/aliases":
		fmt.Fprintf(w, "default")
	case "/computeMetadata/v1/instance/service-accounts/default/email":
		fmt.Fprintf(w, h.sa)
	case "/computeMetadata/v1/instance/service-accounts/default/token":
		scopes := strings.Split(r.URL.Query().Get("scopes"), ",")

		token, err := generateAccessToken(h.sa, scopes)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), 500)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(token)
	case "/computeMetadata/v1/instance/service-accounts/default/identity":
		audience := r.URL.Query().Get("audience")
		token, err := generateIdToken(h.sa, audience)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), 500)
			return
		}
		fmt.Fprintf(w, token)
	case "/computeMetadata/v1/instance/region":
		fmt.Fprintf(w, h.md.Instance.Region)
	case "/computeMetadata/v1/instance/zone":
		fmt.Fprintf(w, "projects/%s/zones/%s-1", h.md.Project.NumericProjectID, h.md.Instance.Region)
	case "/computeMetadata/v1/project/numeric-project-id":
		fmt.Fprintf(w, h.md.Project.NumericProjectID)
	case "/computeMetadata/v1/project/project-id":
		fmt.Fprintf(w, h.md.Project.ProjectID)
	default:
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")
		http.NotFound(w, r)
		return
	}

	return
}
