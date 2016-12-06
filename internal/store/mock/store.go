package mock

type Store struct {
	i      int
	errors []error
	vals   [][]byte
	Puts   map[string]string
}

func NewStore(vals [][]byte, errors []error) *Store {
	return &Store{
		vals:   vals,
		errors: errors,
		Puts:   map[string]string{},
	}
}

func (s *Store) Put(key, val, bucket []byte) error {
	err := s.errors[s.i]
	s.i++
	k := string(key) + "-" + string(bucket)
	s.Puts[k] = string(val)
	return err
}

func (s *Store) Get(key, bucket []byte, f func([]byte) error) error {
	err := s.errors[s.i]
	v := s.vals[s.i]
	if err := f(v); err != nil {
		return err
	}
	s.i++
	return err
}

func (s *Store) GetAll(bucket []byte, f func(val []byte) error) error {
	err := s.errors[s.i]
	s.i++
	return err
}

func (s *Store) Delete(key, bucket []byte) error {
	err := s.errors[s.i]
	s.i++
	return err
}
