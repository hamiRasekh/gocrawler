package crawler

// BuildEmbroideryPayload creates the Elasticsearch payload used for the embroidery
// API crawl. Callers can supply overrides that will be deep-merged into the base
// payload while keeping the paging controls managed by the crawler itself.
func BuildEmbroideryPayload(from, size int, overrides map[string]interface{}) map[string]interface{} {
	payload := defaultEmbroideryPayload(from, size)

	if len(overrides) > 0 {
		mergeEmbroideryOverrides(payload, overrides)
	}

	// Never allow overrides to change paging controls managed by the crawler
	payload["from"] = from
	payload["size"] = size

	return payload
}

func defaultEmbroideryPayload(from, size int) map[string]interface{} {
	return map[string]interface{}{
		"track_total_hits": true,
		"sort": []map[string]string{
			{"saleRank": "desc"},
			{"rating": "desc"},
		},
		"_source": map[string]interface{}{
			"excludes": []string{"*.productTabContent", "*.mainFeatures"},
		},
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []interface{}{},
				"must": []map[string]interface{}{
					{"term": map[string]interface{}{"definitionName": "StockDesign"}},
					{"term": map[string]interface{}{"inStock": true}},
					{"range": map[string]interface{}{"listPrice": map[string]interface{}{"gt": 0}}},
				},
				"must_not": []map[string]interface{}{
					{"term": map[string]interface{}{"definitionName": "PrintArt"}},
					{"term": map[string]interface{}{"definitionName": "SVG"}},
				},
			},
		},
		"from": from,
		"size": size,
		"aggs": map[string]interface{}{
			"Brands": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "brand.raw",
					"order": map[string]string{"_count": "desc"},
					"size":  1000,
				},
			},
			"catalog": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "catalog.raw",
					"order": map[string]string{"_count": "desc"},
					"size":  1000,
				},
			},
			"Artists": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "artist.raw",
					"order": map[string]string{"_count": "desc"},
					"size":  1000,
				},
			},
			"Categories": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "categoriesList.keyword",
					"order": map[string]string{"_count": "desc"},
					"size":  1000,
				},
			},
		},
	}
}

func mergeEmbroideryOverrides(target map[string]interface{}, overrides map[string]interface{}) {
	for key, value := range overrides {
		overrideMap, overrideIsMap := value.(map[string]interface{})

		if existing, ok := target[key]; ok && overrideIsMap {
			if existingMap, ok := existing.(map[string]interface{}); ok {
				mergeEmbroideryOverrides(existingMap, overrideMap)
				continue
			}
		}

		target[key] = value
	}
}

