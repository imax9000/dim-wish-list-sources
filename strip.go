package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

var (
	manifestPath = flag.String("manifest", "manifest.json", "Path to the manifest")
	wishlistPath = flag.String("wishlist", "", "Path to the wishlist")
)

const (
	weaponPerksCategoryHash int64 = 4241085061
	wishlistItemPrefix            = "dimwishlist:"
)

type Manifest struct {
	Items    map[string]Item    `json:"DestinyInventoryItemDefinition"`
	PlugSets map[string]PlugSet `json:"DestinyPlugSetDefinition"`

	perksForItemCache map[string]map[string]bool
}

type Item struct {
	Sockets SocketsDef `json:"sockets"`
}

type SocketsDef struct {
	Entries    []SocketEntry    `json:"socketEntries"`
	Categories []SocketCategory `json:"socketCategories"`
}

type SocketEntry struct {
	PlugSetHash int64 `json:"randomizedPlugSetHash"`
}

type SocketCategory struct {
	Hash    int64 `json:"socketCategoryHash"`
	Indexes []int `json:"socketIndexes"`
}

type PlugSet struct {
	Items []PlugItem `json:"reusablePlugItems"`
}

type PlugItem struct {
	Hash int64 `json:"plugItemHash"`
}

func (m *Manifest) PerksForItem(hash string) (map[string]bool, error) {
	if r := m.perksForItemCache[hash]; r != nil {
		return r, nil
	}

	item, found := m.Items[hash]
	if !found {
		return nil, fmt.Errorf("item %q not found", hash)
	}

	plugSets, err := item.PlugSetHashes()
	if err != nil {
		return nil, fmt.Errorf("item %q: %w", hash, err)
	}

	ret := map[string]bool{}
	for _, p := range plugSets {
		plugSet, found := m.PlugSets[strconv.FormatInt(p, 10)]
		if !found {
			return nil, fmt.Errorf("item %q: plug set %d not found", hash, p)
		}
		for _, pi := range plugSet.Items {
			ret[strconv.FormatInt(pi.Hash, 10)] = true
		}
	}

	if m.perksForItemCache == nil {
		m.perksForItemCache = map[string]map[string]bool{}
	}
	m.perksForItemCache[hash] = ret
	return ret, nil
}

func (i Item) SocketCategory(hash int64) (SocketCategory, error) {
	for _, c := range i.Sockets.Categories {
		if c.Hash == hash {
			return c, nil
		}
	}
	return SocketCategory{}, fmt.Errorf("socket category %d not found", hash)
}

func (i Item) PlugSetHashes() ([]int64, error) {
	cat, err := i.SocketCategory(weaponPerksCategoryHash)
	if err != nil {
		return nil, err
	}
	if len(cat.Indexes) < 4 {
		return nil, fmt.Errorf("weapon perks category has %d sockets, want at leas 4", len(cat.Indexes))
	}
	ret := []int64{}
	for _, index := range []int{cat.Indexes[2], cat.Indexes[3]} {
		if index > len(i.Sockets.Entries) {
			return nil, fmt.Errorf("socket entry with index %d not found, total entries: %d", index, len(i.Sockets.Entries))
		}
		ret = append(ret, i.Sockets.Entries[index].PlugSetHash)
	}
	return ret, nil
}

func filterWishlist(wishlist io.Reader, out io.Writer, manifest *Manifest) error {
	scanner := bufio.NewScanner(wishlist)
	lines := []string{}
	process := func(lines []string) {
		filtered, err := filterItemSet(lines, manifest)
		if err != nil {
			log.Println(err)
			filtered = lines // So we write out the original entries
		}
		for _, l := range filtered {
			fmt.Fprintln(out, l)
		}
	}
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, wishlistItemPrefix) {
			lines = append(lines, line)
			continue
		}

		// We're at a non-item line, process what we have accumulated so far.
		if len(lines) > 0 {
			process(lines)
			lines = nil
		}

		fmt.Fprintln(out, line)
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading wishlist: %w", err)
	}

	// Process the last set in the file.
	if len(lines) > 0 {
		process(lines)
	}
	return nil
}

type wishlistEntry struct {
	Item  string
	Perks []string
	Notes string
}

func (e wishlistEntry) String() string {
	r := fmt.Sprintf("%sitem=%s&perks=%s", wishlistItemPrefix, e.Item, strings.Join(e.Perks, ","))
	if e.Notes != "" {
		r += "#notes:" + e.Notes
	}
	return r
}

func filterItemSet(lines []string, manifest *Manifest) ([]string, error) {
	entries := map[string][]wishlistEntry{}
	for _, line := range lines {
		entry, err := parseWishlistEntry(line)
		if err != nil {
			return nil, fmt.Errorf("while parsing %q: %w", line, err)
		}
		perks, err := manifest.PerksForItem(entry.Item)
		if err != nil {
			return nil, fmt.Errorf("while looking up perks for the item in %q: %w", line, err)
		}
		filtered := []string{}
		for _, p := range entry.Perks {
			if perks[p] {
				filtered = append(filtered, p)
			}
		}
		entry.Perks = filtered
		if len(entry.Perks) > 0 {
			entries[entry.Item] = append(entries[entry.Item], entry)
		}
	}

	ret := []string{}
	for _, entries := range entries {
		for _, entry := range dedupeByPerkSets(entries) {
			ret = append(ret, entry.String())
		}
	}
	return ret, nil
}

func parseWishlistEntry(line string) (wishlistEntry, error) {
	ret := wishlistEntry{}

	parts := strings.SplitN(strings.TrimPrefix(line, wishlistItemPrefix), "#notes:", 2)
	if len(parts) >= 2 {
		ret.Notes = parts[1]
	}

	fields := strings.Split(parts[0], "&")
	for _, field := range fields {
		parts := strings.SplitN(field, "=", 2)
		if len(parts) != 2 {
			return wishlistEntry{}, fmt.Errorf("missing \"=\" in %q", field)
		}
		name := parts[0]
		value := parts[1]
		switch name {
		case "item":
			ret.Item = value
		case "perks":
			ret.Perks = strings.Split(value, ",")
		default:
			return wishlistEntry{}, fmt.Errorf("unknown field %q", name)
		}
	}
	if ret.Item == "" {
		return wishlistEntry{}, fmt.Errorf("missing item ID")
	}
	if ret.Perks == nil {
		return wishlistEntry{}, fmt.Errorf("missing perks")
	}

	return ret, nil
}

func dedupeByPerkSets(entries []wishlistEntry) []wishlistEntry {
	ret := []wishlistEntry{}
	sets := map[string]bool{}
	notes := map[string]map[string]bool{}

	perkSet := func(entry wishlistEntry) string {
		p := make([]string, len(entry.Perks))
		copy(p, entry.Perks)
		sort.Strings(p)
		return strings.Join(p, ",")
	}

	for _, entry := range entries {
		s := perkSet(entry)
		if sets[s] {
			notes[s][entry.Notes] = true
			continue
		}

		sets[s] = true
		ret = append(ret, entry)
		notes[s] = map[string]bool{entry.Notes: true}
	}

	for i, entry := range ret {
		n := []string{}
		for note := range notes[perkSet(entry)] {
			if note == "" {
				continue
			}
			n = append(n, note)
		}
		sort.Strings(n)
		entry.Notes = strings.Join(n, ";")
		ret[i] = entry
	}
	return ret
}

func main() {
	flag.Parse()
	m, err := os.Open(*manifestPath)
	if err != nil {
		log.Fatalf("failed to open the manifest: %s", err)
	}
	wishlist := os.Stdin
	if *wishlistPath != "" {
		wishlist, err = os.Open(*wishlistPath)
		if err != nil {
			log.Fatalf("failed to open the wishlist: %s", err)
		}
	}
	manifest := &Manifest{}
	if err := json.NewDecoder(m).Decode(manifest); err != nil {
		log.Fatalf("failed to parse the manifest: %s", err)
	}
	if err := filterWishlist(wishlist, os.Stdout, manifest); err != nil {
		log.Fatal(err)
	}
}
