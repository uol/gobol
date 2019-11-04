package solar

import (
	"errors"
	"fmt"
	"net/url"
	"sync"

	"github.com/uol/go-solr/solr"
	"github.com/uol/gobol/logh"
)

/**
* Contains all main structs and functions.
* @author rnojiri
**/

// SolrService - struct
type SolrService struct {
	solrCollectionsAdmin *solr.CollectionsAdmin
	loggers              *logh.ContextualLogger
	url                  string
	solrInterfaceCache   sync.Map
}

// recoverFromFailure - recovers from a failure
func (ss *SolrService) recoverFromFailure(funcName string) {
	if r := recover(); r != nil {
		if logh.ErrorEnabled {
			ss.loggers.Error().Str("func", funcName).Msg(fmt.Sprintf("recovered from: %s", r))
		}
	}
}

// NewSolrService - creates a new instance
func NewSolrService(url string) (*SolrService, error) {

	sca, err := solr.NewCollectionsAdmin(url)
	if err != nil {
		return nil, err
	}

	return &SolrService{
		solrCollectionsAdmin: sca,
		loggers:              logh.CreateContextualLogger("pkg", "solar/solr_service"),
		url:                  url,
		solrInterfaceCache:   sync.Map{},
	}, nil
}

// getSolrInterface - creates a new solr interface based on the given collection
func (ss *SolrService) getSolrInterface(collection string) (*solr.SolrInterface, error) {

	if si, ok := ss.solrInterfaceCache.Load(collection); ok {
		return si.(*solr.SolrInterface), nil
	}

	si, err := solr.NewSolrInterface(ss.url, collection)
	if err != nil {
		if logh.ErrorEnabled {
			ss.loggers.Error().Msg("error creating a new instance of solr interface")
		}
		return nil, err
	}

	ss.solrInterfaceCache.Store(collection, si)

	return si, err
}

// AddDocument - add one document to the solr collection
func (ss *SolrService) AddDocument(collection string, commit bool, doc *solr.Document) error {

	defer ss.recoverFromFailure("AddDocuments")

	if doc == nil {
		return errors.New("document is null")
	}

	si, err := ss.getSolrInterface(collection)
	if err != nil {
		if logh.ErrorEnabled {
			ss.loggers.Error().Msg("error getting solr interface")
		}
		return err
	}

	params := &url.Values{}
	if commit {
		params.Add("commit", "true")
	}

	_, err = si.Add([]solr.Document{*doc}, 0, params)
	if err != nil {
		if logh.ErrorEnabled {
			ss.loggers.Error().Msg(fmt.Sprintf("error adding 1 document to the collection %s: %s", collection, err.Error()))
		}
		return err
	}

	if logh.InfoEnabled {
		ss.loggers.Info().Msg(fmt.Sprintf("added 1 documents to the collection %s", collection))
	}

	return nil
}

// AddDocuments - add one or more documentos to the solr collection
func (ss *SolrService) AddDocuments(collection string, commit bool, docs ...solr.Document) error {

	defer ss.recoverFromFailure("AddDocuments")

	if docs == nil || len(docs) == 0 {
		return errors.New("no documents to add")
	}

	si, err := ss.getSolrInterface(collection)
	if err != nil {
		if logh.ErrorEnabled {
			ss.loggers.Error().Msg("error getting solr interface")
		}
		return err
	}

	params := &url.Values{}
	if commit {
		params.Add("commit", "true")
	}

	numDocs := len(docs)

	_, err = si.Add(docs, 0, params)
	if err != nil {
		if logh.ErrorEnabled {
			ss.loggers.Error().Msg(fmt.Sprintf("error adding %d document to the collection %s: %s", numDocs, collection, err.Error()))
		}
		return err
	}

	if logh.InfoEnabled {
		ss.loggers.Info().Msg(fmt.Sprintf("added %d documents to the collection %s", numDocs, collection))
	}

	return nil
}

// DeleteDocumentByID - delete a document by ID
func (ss *SolrService) DeleteDocumentByID(collection string, commit bool, id string) error {

	defer ss.recoverFromFailure("DeleteDocumentByID")

	if id == "" {
		return errors.New("document id not informed, no document will be deleted")
	}

	query := fmt.Sprintf("id:%s", id)

	err := ss.DeleteDocumentByQuery(collection, commit, query)
	if err != nil {
		if logh.ErrorEnabled {
			ss.loggers.Error().Msg(fmt.Sprintf("error deleting document %s of collection %s: %s", id, collection, err.Error()))
		}
		return err
	}

	return nil
}

// DeleteDocumentByQuery - delete document by query
func (ss *SolrService) DeleteDocumentByQuery(collection string, commit bool, query string) error {

	defer ss.recoverFromFailure("DeleteDocumentByQuery")

	if query == "" {
		return errors.New("query not informed, no document will be deleted")
	}

	si, err := ss.getSolrInterface(collection)
	if err != nil {
		if logh.ErrorEnabled {
			ss.loggers.Error().Msg("error getting solr interface")
		}
		return err
	}

	params := &url.Values{}
	if commit {
		params.Add("commit", "true")
	}

	doc := map[string]interface{}{}
	doc["query"] = query

	_, err = si.Delete(doc, params)
	if err != nil {
		if logh.ErrorEnabled {
			ss.loggers.Error().Msg(fmt.Sprintf("error deleting document of collection %s with query %s: %s", collection, query, err.Error()))
		}
		return err
	}

	if logh.InfoEnabled {
		ss.loggers.Info().Msg(fmt.Sprintf("deleted document(s) of collection %s with query %s", collection, query))
	}

	return nil
}
