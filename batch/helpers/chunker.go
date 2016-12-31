package helpers

import "github.com/kbuzsaki/wikidegree/wiki"

func ChunkPageBuffers(chunkSize int, pageBuffers <-chan []wiki.Page, chunkedBuffers chan<- []wiki.Page) {
	for pageBuffer := range pageBuffers {
		for start := 0; start < len(pageBuffer); start += chunkSize {
			end := start + chunkSize
			if end > len(pageBuffer) {
				end = len(pageBuffer)
			}

			chunk := pageBuffer[start:end]
			chunkedBuffers <- chunk
		}
	}

	close(chunkedBuffers)
}
