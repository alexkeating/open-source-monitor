package lib

import "io/ioutil"

func OpenFile(path string) ([]byte, error) {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return contents, nil
}
