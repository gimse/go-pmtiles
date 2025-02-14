package pmtiles

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

func TestDirectoryRoundtrip(t *testing.T) {
	entries := make([]EntryV3, 0)
	entries = append(entries, EntryV3{0, 0, 0, 0})
	entries = append(entries, EntryV3{1, 1, 1, 1})
	entries = append(entries, EntryV3{2, 2, 2, 2})

	serialized := serialize_entries(entries)
	result := deserialize_entries(bytes.NewBuffer(serialized))
	assert.Equal(t, 3, len(result))
	assert.Equal(t, uint64(0), result[0].TileId)
	assert.Equal(t, uint64(0), result[0].Offset)
	assert.Equal(t, uint32(0), result[0].Length)
	assert.Equal(t, uint32(0), result[0].RunLength)
	assert.Equal(t, uint64(1), result[1].TileId)
	assert.Equal(t, uint64(1), result[1].Offset)
	assert.Equal(t, uint32(1), result[1].Length)
	assert.Equal(t, uint32(1), result[1].RunLength)
	assert.Equal(t, uint64(2), result[2].TileId)
	assert.Equal(t, uint64(2), result[2].Offset)
	assert.Equal(t, uint32(2), result[2].Length)
	assert.Equal(t, uint32(2), result[2].RunLength)
}

func TestHeaderRoundtrip(t *testing.T) {
	header := HeaderV3{}
	header.RootOffset = 1
	header.RootLength = 2
	header.MetadataOffset = 3
	header.MetadataLength = 4
	header.LeafDirectoryOffset = 5
	header.LeafDirectoryLength = 6
	header.TileDataOffset = 7
	header.TileDataLength = 8
	header.AddressedTilesCount = 9
	header.TileEntriesCount = 10
	header.TileContentsCount = 11
	header.Clustered = true
	header.InternalCompression = Gzip
	header.TileCompression = Brotli
	header.TileType = Mvt
	header.MinZoom = 1
	header.MaxZoom = 2
	header.MinLonE7 = 1.1 * 10000000
	header.MinLatE7 = 2.1 * 10000000
	header.MaxLonE7 = 1.2 * 10000000
	header.MaxLatE7 = 2.2 * 10000000
	header.CenterZoom = 3
	header.CenterLonE7 = 3.1 * 10000000
	header.CenterLatE7 = 3.2 * 10000000
	b := serialize_header(header)
	result, _ := deserialize_header(b)
	assert.Equal(t, uint64(1), result.RootOffset)
	assert.Equal(t, uint64(2), result.RootLength)
	assert.Equal(t, uint64(3), result.MetadataOffset)
	assert.Equal(t, uint64(4), result.MetadataLength)
	assert.Equal(t, uint64(5), result.LeafDirectoryOffset)
	assert.Equal(t, uint64(6), result.LeafDirectoryLength)
	assert.Equal(t, uint64(7), result.TileDataOffset)
	assert.Equal(t, uint64(8), result.TileDataLength)
	assert.Equal(t, uint64(9), result.AddressedTilesCount)
	assert.Equal(t, uint64(10), result.TileEntriesCount)
	assert.Equal(t, uint64(11), result.TileContentsCount)
	assert.Equal(t, true, result.Clustered)
	assert.Equal(t, Gzip, int(result.InternalCompression))
	assert.Equal(t, Brotli, int(result.TileCompression))
	assert.Equal(t, Mvt, int(result.TileType))
	assert.Equal(t, uint8(1), result.MinZoom)
	assert.Equal(t, uint8(2), result.MaxZoom)
	assert.Equal(t, int32(11000000), result.MinLonE7)
	assert.Equal(t, int32(21000000), result.MinLatE7)
	assert.Equal(t, int32(12000000), result.MaxLonE7)
	assert.Equal(t, int32(22000000), result.MaxLatE7)
	assert.Equal(t, uint8(3), result.CenterZoom)
	assert.Equal(t, int32(31000000), result.CenterLonE7)
	assert.Equal(t, int32(32000000), result.CenterLatE7)
}

func TestOptimizeDirectories(t *testing.T) {
	rand.Seed(3857)
	entries := make([]EntryV3, 0)
	entries = append(entries, EntryV3{0, 0, 100, 1})
	_, leaves_bytes, num_leaves := optimize_directories(entries, 100)
	assert.False(t, len(leaves_bytes) > 0)
	assert.Equal(t, 0, num_leaves)

	entries = make([]EntryV3, 0)
	var i uint64
	var offset uint64
	for ; i < 1000; i++ {
		randtilesize := rand.Intn(1000000)
		entries = append(entries, EntryV3{i, offset, uint32(randtilesize), 1})
		offset += uint64(randtilesize)
	}

	root_bytes, leaves_bytes, num_leaves := optimize_directories(entries, 1024)

	assert.False(t, len(root_bytes) > 1024)

	assert.False(t, num_leaves == 0)
	assert.False(t, len(leaves_bytes) == 0)
}

func TestFindTileMissing(t *testing.T) {
	entries := make([]EntryV3, 0)
	_, ok := find_tile(entries, 0)
	assert.False(t, ok)
}

func TestFindTileFirstEntry(t *testing.T) {
	entries := []EntryV3{{TileId: 100, Offset: 1, Length: 1, RunLength: 1}}
	entry, ok := find_tile(entries, 100)
	assert.Equal(t, true, ok)
	assert.Equal(t, uint64(1), entry.Offset)
	assert.Equal(t, uint32(1), entry.Length)
	_, ok = find_tile(entries, 101)
	assert.Equal(t, false, ok)
}

func TestFindTileMultipleEntries(t *testing.T) {
	entries := []EntryV3{
		{TileId: 100, Offset: 1, Length: 1, RunLength: 2},
	}
	entry, ok := find_tile(entries, 101)
	assert.Equal(t, true, ok)
	assert.Equal(t, uint64(1), entry.Offset)
	assert.Equal(t, uint32(1), entry.Length)

	entries = []EntryV3{
		{TileId: 100, Offset: 1, Length: 1, RunLength: 1},
		{TileId: 150, Offset: 2, Length: 2, RunLength: 2},
	}
	entry, ok = find_tile(entries, 151)
	assert.Equal(t, true, ok)
	assert.Equal(t, uint64(2), entry.Offset)
	assert.Equal(t, uint32(2), entry.Length)

	entries = []EntryV3{
		{TileId: 50, Offset: 1, Length: 1, RunLength: 2},
		{TileId: 100, Offset: 2, Length: 2, RunLength: 1},
		{TileId: 150, Offset: 3, Length: 3, RunLength: 1},
	}
	entry, ok = find_tile(entries, 51)
	assert.Equal(t, true, ok)
	assert.Equal(t, uint64(1), entry.Offset)
	assert.Equal(t, uint32(1), entry.Length)
}

func TestFindTileLeafSearch(t *testing.T) {
	entries := []EntryV3{
		{TileId: 100, Offset: 1, Length: 1, RunLength: 0},
	}
	entry, ok := find_tile(entries, 150)
	assert.Equal(t, true, ok)
	assert.Equal(t, uint64(1), entry.Offset)
	assert.Equal(t, uint32(1), entry.Length)
}

func TestBuildRootsLeaves(t *testing.T) {
	entries := []EntryV3{
		{TileId: 100, Offset: 1, Length: 1, RunLength: 0},
	}
	_, _, num_leaves := build_roots_leaves(entries, 1)
	assert.Equal(t, 1, num_leaves)
}
