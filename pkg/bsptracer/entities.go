package bsptracer

import "strings"

func parseEntities(str string) []map[string]string {
	blocks := strings.Split(str, "}")
	entities := make([]map[string]string, 0, len(blocks))

	for _, block := range blocks {
		block = strings.TrimPrefix(strings.TrimSpace(block), "{")
		kvEntry := strings.Split(block, "\n")

		data := make(map[string]string)

		for _, entry := range kvEntry {
			kv := strings.Split(entry, "\"")
			if len(kv) != 4 {
				continue
			}

			k := kv[1]
			v := kv[3]

			data[k] = v
		}

		entities = append(entities, data)
	}

	return entities
}
