package ext

// This file declares a Regexp struct that wraps around regexp.Regexp.
// The wrapper implements marshalling functions so that you can serialize and
// deserialize regular expressions from JSON.

import "regexp"

type Regexp struct {
	*regexp.Regexp
}

func (re *Regexp) UnmarshalText(b []byte) error {
	regex, err := regexp.Compile(string(b))
	if err != nil {
		return err
	}
	re.Regexp = regex
	return nil
}

func (re *Regexp) MarshalText() ([]byte, error) {
	if re.Regexp != nil {
		return []byte(re.Regexp.String()), nil
	}
	return nil, nil
}
