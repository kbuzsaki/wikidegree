# Six Degrees of Wikipedia

This is an implementation of a [six degrees of wikipedia](https://en.wikipedia.org/wiki/Wikipedia:Six_degrees_of_Wikipedia)
solver written in go. It uses an offline copy the full text of wikipedia (52 gigabytes!) as its basis, and builds
an index of page relationships using the [bolt](https://github.com/boltdb/bolt) key/value store. This final index takes
up about 20 gigabytes on disk, though much of that is empty buffer space.

You can find the web client for it running at https://wikidegree.kbuzsaki.com
