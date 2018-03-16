package mgobson_test

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mongodb-labs/mgobson"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/stretchr/testify/require"
)

func TestAll(t *testing.T) {
	t.Run("mgobson.M", func(t *testing.T) {
		t.Run("MarshalBSONDocument", func(t *testing.T) {
			testCases := []struct {
				name string
				m    mgobson.M
				doc  *bson.Document
				err  error
			}{
				{
					"empty",
					mgobson.M{},
					bson.NewDocument(),
					nil,
				},
				{
					"simple",
					mgobson.M{
						"foo": int32(1),
						"bar": false,
					},
					bson.NewDocument(
						bson.EC.Int32("foo", 1),
						bson.EC.Boolean("bar", false),
					),
					nil,
				},
				{
					"nested",
					mgobson.M{
						"foo": mgobson.M{
							"bar": false,
						},
					},
					bson.NewDocument(
						bson.EC.SubDocumentFromElements("foo",
							bson.EC.Boolean("bar", false),
						),
					),
					nil,
				},
			}

			for _, tc := range testCases {
				doc, err := tc.m.MarshalBSONDocument()
				require.Equal(t, tc.err, err)
				if err != nil {
					continue
				}

				require.True(t, cmp.Equal(tc.doc, doc))
			}
		})

		t.Run("Marshal", func(t *testing.T) {
			testCases := []struct {
				name string
				m    mgobson.M
				b    []byte
				err  error
			}{
				{
					"empty",
					mgobson.M{},
					[]byte{5, 0, 0, 0, 0},
					nil,
				},
				{
					"simple",
					mgobson.M{
						"foo": int32(1),
						"bar": false,
					},
					[]byte{
						// length - 20
						0x14, 0x0, 0x0, 0x0,

						// type - int32
						0x10,
						// key - "foo"
						0x66, 0x6f, 0x6f, 0x0,
						// value - int32(1)
						0x1, 0x0, 0x0, 0x0,

						// type - bool
						0x8,
						// key - "bar"
						0x62, 0x61, 0x72, 0x0,
						// value - false
						0x0,

						// null terminator
						0x0,
					},
					nil,
				},
				{
					"nested",
					mgobson.M{
						"foo": mgobson.M{
							"bar": false,
						},
					},
					[]byte{
						// length - 20
						0x15, 0x0, 0x0, 0x0,

						// type - document
						0x3,
						// key - "foo"
						0x66, 0x6f, 0x6f, 0x0,

						// ----- begin subdocument -----

						// length - xx
						0xb, 0x0, 0x0, 0x0,

						// type - bool
						0x8,
						// key - "bar"
						0x62, 0x61, 0x72, 0x0,
						// value - false
						0x0,

						// null terminator
						0x0,

						// ----- end subdocument -----

						// null terminator
						0x0,
					},
					nil,
				},
			}

			for _, tc := range testCases {
				b, err := tc.m.MarshalBSON()
				require.Equal(t, tc.err, err)
				if err != nil {
					continue
				}

				require.True(t, bytes.Equal(tc.b, b))
			}
		})

		t.Run("Unmarshal", func(t *testing.T) {
			testCases := []struct {
				name     string
				actual   mgobson.M
				b        []byte
				expected mgobson.M
				err      error
			}{
				{
					"empty",
					mgobson.M{},
					[]byte{5, 0, 0, 0, 0},
					mgobson.M{},
					nil,
				},
				{
					"simple",
					mgobson.M{},
					[]byte{
						// length - 20
						0x14, 0x0, 0x0, 0x0,

						// type - int32
						0x10,
						// key - "foo"
						0x66, 0x6f, 0x6f, 0x0,
						// value - int32(1)
						0x1, 0x0, 0x0, 0x0,

						// type - bool
						0x8,
						// key - "bar"
						0x62, 0x61, 0x72, 0x0,
						// value - false
						0x0,

						// null terminator
						0x0,
					},
					mgobson.M{
						"foo": int32(1),
						"bar": false,
					},
					nil,
				},
				{
					"nested",
					mgobson.M{},
					[]byte{
						// length - 20
						0x15, 0x0, 0x0, 0x0,

						// type - document
						0x3,
						// key - "foo"
						0x66, 0x6f, 0x6f, 0x0,

						// ----- begin subdocument -----

						// length - xx
						0xb, 0x0, 0x0, 0x0,

						// type - bool
						0x8,
						// key - "bar"
						0x62, 0x61, 0x72, 0x0,
						// value - false
						0x0,

						// null terminator
						0x0,

						// ----- end subdocument -----

						// null terminator
						0x0,
					},
					mgobson.M{
						"foo": mgobson.M{
							"bar": false,
						},
					},
					nil,
				},
			}

			for _, tc := range testCases {
				err := tc.actual.UnmarshalBSON(tc.b)
				require.Equal(t, tc.err, err)
				if err != nil {
					continue
				}

				require.True(t, cmp.Equal(tc.actual, tc.expected))
			}
		})
	})

	t.Run("mgobson.D", func(t *testing.T) {
		t.Run("MarshalBSONDocument", func(t *testing.T) {
			testCases := []struct {
				name string
				d    mgobson.D
				doc  *bson.Document
				err  error
			}{
				{
					"empty",
					mgobson.D{},
					bson.NewDocument(),
					nil,
				},
				{
					"simple",
					mgobson.D{
						{"foo", int32(1)},
						{"bar", false},
					},
					bson.NewDocument(
						bson.EC.Int32("foo", 1),
						bson.EC.Boolean("bar", false),
					),
					nil,
				},
				{
					"nested",
					mgobson.D{
						{"foo", mgobson.D{
							{"bar", false},
						}},
					},
					bson.NewDocument(
						bson.EC.SubDocumentFromElements("foo",
							bson.EC.Boolean("bar", false),
						),
					),
					nil,
				},
			}

			for _, tc := range testCases {
				doc, err := tc.d.MarshalBSONDocument()
				require.Equal(t, tc.err, err)
				if err != nil {
					continue
				}

				require.True(t, cmp.Equal(tc.doc, doc))
			}
		})

		t.Run("Marshal", func(t *testing.T) {
			testCases := []struct {
				name string
				d    mgobson.D
				b    []byte
				err  error
			}{
				{
					"empty",
					mgobson.D{},
					[]byte{5, 0, 0, 0, 0},
					nil,
				},
				{
					"simple",
					mgobson.D{
						{"foo", int32(1)},
						{"bar", false},
					},
					[]byte{
						// length - 20
						0x14, 0x0, 0x0, 0x0,

						// type - int32
						0x10,
						// key - "foo"
						0x66, 0x6f, 0x6f, 0x0,
						// value - int32(1)
						0x1, 0x0, 0x0, 0x0,

						// type - bool
						0x8,
						// key - "bar"
						0x62, 0x61, 0x72, 0x0,
						// value - false
						0x0,

						// null terminator
						0x0,
					},
					nil,
				},
				{
					"nested",
					mgobson.D{
						{"foo", mgobson.D{
							{"bar", false},
						}},
					},
					[]byte{
						// length - 20
						0x15, 0x0, 0x0, 0x0,

						// type - document
						0x3,
						// key - "foo"
						0x66, 0x6f, 0x6f, 0x0,

						// ----- begin subdocument -----

						// length - xx
						0xb, 0x0, 0x0, 0x0,

						// type - bool
						0x8,
						// key - "bar"
						0x62, 0x61, 0x72, 0x0,
						// value - false
						0x0,

						// null terminator
						0x0,

						// ----- end subdocument -----

						// null terminator
						0x0,
					},
					nil,
				},
			}

			for _, tc := range testCases {
				b, err := tc.d.MarshalBSON()
				require.Equal(t, tc.err, err)
				if err != nil {
					continue
				}

				require.True(t, bytes.Equal(tc.b, b))
			}
		})

		t.Run("Unmarshal", func(t *testing.T) {
			testCases := []struct {
				name     string
				actual   mgobson.D
				b        []byte
				expected mgobson.D
				err      error
			}{
				{
					"empty",
					mgobson.D{},
					[]byte{5, 0, 0, 0, 0},
					mgobson.D{},
					nil,
				},
				{
					"simple",
					mgobson.D{},
					[]byte{
						// length - 20
						0x14, 0x0, 0x0, 0x0,

						// type - int32
						0x10,
						// key - "foo"
						0x66, 0x6f, 0x6f, 0x0,
						// value - int32(1)
						0x1, 0x0, 0x0, 0x0,

						// type - bool
						0x8,
						// key - "bar"
						0x62, 0x61, 0x72, 0x0,
						// value - false
						0x0,

						// null terminator
						0x0,
					},
					mgobson.D{
						{"foo", int32(1)},
						{"bar", false},
					},
					nil,
				},
				{
					"nested",
					mgobson.D{},
					[]byte{
						// length - 20
						0x15, 0x0, 0x0, 0x0,

						// type - document
						0x3,
						// key - "foo"
						0x66, 0x6f, 0x6f, 0x0,

						// ----- begin subdocument -----

						// length - xx
						0xb, 0x0, 0x0, 0x0,

						// type - bool
						0x8,
						// key - "bar"
						0x62, 0x61, 0x72, 0x0,
						// value - false
						0x0,

						// null terminator
						0x0,

						// ----- end subdocument -----

						// null terminator
						0x0,
					},
					mgobson.D{
						{"foo", mgobson.D{
							{"bar", false},
						}},
					},
					nil,
				},
			}

			for _, tc := range testCases {
				err := tc.actual.UnmarshalBSON(tc.b)
				require.Equal(t, tc.err, err)
				if err != nil {
					continue
				}

				require.True(t, cmp.Equal(tc.actual, tc.expected))
			}
		})
	})

	t.Run("mgobson.RawD", func(t *testing.T) {
		t.Run("MarshalBSONDocument", func(t *testing.T) {
			testCases := []struct {
				name string
				d    mgobson.RawD
				doc  *bson.Document
				err  error
			}{
				{
					"empty",
					mgobson.RawD{},
					bson.NewDocument(),
					nil,
				},
				{
					"simple",
					mgobson.RawD{
						{
							"foo",
							mgobson.Raw{
								Kind: 0x10,
								Data: []byte{
									// int32(1)
									0x1, 0x0, 0x0, 0x0,
								},
							},
						},
						{
							"bar",
							mgobson.Raw{
								Kind: 0x8,
								Data: []byte{
									// false
									0x0,
								},
							},
						},
					},
					bson.NewDocument(
						bson.EC.Int32("foo", 1),
						bson.EC.Boolean("bar", false),
					),
					nil,
				},
				{
					"nested",
					mgobson.RawD{
						{
							"foo",
							mgobson.Raw{
								Kind: 0x3,
								Data: []byte{
									// length - xx
									0xb, 0x0, 0x0, 0x0,

									// type - bool
									0x8,
									// key - "bar"
									0x62, 0x61, 0x72, 0x0,
									// value - false
									0x0,

									// null terminator
									0x0,
								},
							},
						},
					},
					bson.NewDocument(
						bson.EC.SubDocumentFromElements("foo",
							bson.EC.Boolean("bar", false),
						),
					),
					nil,
				},
			}

			for _, tc := range testCases {
				doc, err := tc.d.MarshalBSONDocument()
				require.Equal(t, tc.err, err)
				if err != nil {
					continue
				}

				require.True(t, cmp.Equal(tc.doc, doc))
			}
		})

		t.Run("Marshal", func(t *testing.T) {
			testCases := []struct {
				name string
				r    mgobson.RawD
				b    []byte
				err  error
			}{
				{
					"empty",
					mgobson.RawD{},
					[]byte{5, 0, 0, 0, 0},
					nil,
				},
				{
					"simple",
					mgobson.RawD{
						{
							"foo",
							mgobson.Raw{
								Kind: 0x10,
								Data: []byte{
									// int32(1)
									0x1, 0x0, 0x0, 0x0,
								},
							},
						},
						{
							"bar",
							mgobson.Raw{
								Kind: 0x8,
								Data: []byte{
									// false
									0x0,
								},
							},
						},
					},
					[]byte{
						// length - 20
						0x14, 0x0, 0x0, 0x0,

						// type - int32
						0x10,
						// key - "foo"
						0x66, 0x6f, 0x6f, 0x0,
						// value - int32(1)
						0x1, 0x0, 0x0, 0x0,

						// type - bool
						0x8,
						// key - "bar"
						0x62, 0x61, 0x72, 0x0,
						// value - false
						0x0,

						// null terminator
						0x0,
					},
					nil,
				},
				{
					"nested",
					mgobson.RawD{
						{
							"foo",
							mgobson.Raw{
								Kind: 0x3,
								Data: []byte{
									// length - xx
									0xb, 0x0, 0x0, 0x0,

									// type - bool
									0x8,
									// key - "bar"
									0x62, 0x61, 0x72, 0x0,
									// value - false
									0x0,

									// null terminator
									0x0,
								},
							},
						},
					},
					[]byte{
						// length - 20
						0x15, 0x0, 0x0, 0x0,

						// type - document
						0x3,
						// key - "foo"
						0x66, 0x6f, 0x6f, 0x0,

						// ----- begin subdocument -----

						// length - xx
						0xb, 0x0, 0x0, 0x0,

						// type - bool
						0x8,
						// key - "bar"
						0x62, 0x61, 0x72, 0x0,
						// value - false
						0x0,

						// null terminator
						0x0,

						// ----- end subdocument -----

						// null terminator
						0x0,
					},
					nil,
				},
			}

			for _, tc := range testCases {
				b, err := tc.r.MarshalBSON()
				require.Equal(t, tc.err, err)
				if err != nil {
					continue
				}

				require.True(t, bytes.Equal(tc.b, b))
			}
		})

		t.Run("Unmarshal", func(t *testing.T) {
			testCases := []struct {
				name     string
				actual   mgobson.RawD
				b        []byte
				expected mgobson.RawD
				err      error
			}{
				{
					"empty",
					mgobson.RawD{},
					[]byte{5, 0, 0, 0, 0},
					mgobson.RawD{},
					nil,
				},
				{
					"simple",
					mgobson.RawD{},
					[]byte{
						// length - 20
						0x14, 0x0, 0x0, 0x0,

						// type - int32
						0x10,
						// key - "foo"
						0x66, 0x6f, 0x6f, 0x0,
						// value - int32(1)
						0x1, 0x0, 0x0, 0x0,

						// type - bool
						0x8,
						// key - "bar"
						0x62, 0x61, 0x72, 0x0,
						// value - false
						0x0,

						// null terminator
						0x0,
					},
					mgobson.RawD{
						{
							"foo",
							mgobson.Raw{
								Kind: 0x10,
								Data: []byte{
									// int32(1)
									0x1, 0x0, 0x0, 0x0,
								},
							},
						},
						{
							"bar",
							mgobson.Raw{
								Kind: 0x8,
								Data: []byte{
									// false
									0x0,
								},
							},
						},
					},
					nil,
				},
				{
					"nested",
					mgobson.RawD{},
					[]byte{
						// length - 20
						0x15, 0x0, 0x0, 0x0,

						// type - document
						0x3,
						// key - "foo"
						0x66, 0x6f, 0x6f, 0x0,

						// ----- begin subdocument -----

						// length - xx
						0xb, 0x0, 0x0, 0x0,

						// type - bool
						0x8,
						// key - "bar"
						0x62, 0x61, 0x72, 0x0,
						// value - false
						0x0,

						// null terminator
						0x0,

						// ----- end subdocument -----

						// null terminator
						0x0,
					},
					mgobson.RawD{
						{
							"foo",
							mgobson.Raw{
								Kind: 0x3,
								Data: []byte{
									// length - xx
									0xb, 0x0, 0x0, 0x0,

									// type - bool
									0x8,
									// key - "bar"
									0x62, 0x61, 0x72, 0x0,
									// value - false
									0x0,

									// null terminator
									0x0,
								},
							},
						},
					},
					nil,
				},
			}

			for _, tc := range testCases {
				err := tc.actual.UnmarshalBSON(tc.b)
				require.Equal(t, tc.err, err)
				if err != nil {
					continue
				}

				require.True(t, cmp.Equal(tc.actual, tc.expected))
			}
		})
	})
}
