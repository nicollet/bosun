package dbtest

import (
	"testing"
	"time"

	"bosun.org/models"
)

func TestSilence(t *testing.T) {
	sd := testData.Silence()

	silence := &models.Silence{
		Start: time.Now().Add(-48 * time.Hour),
		End:   time.Now().Add(5 * time.Hour),
		Alert: "Foo",
	}
	future := &models.Silence{
		Start: time.Now().Add(1 * time.Hour),
		End:   time.Now().Add(2 * time.Hour),
		Alert: "Foo",
	}
	past := &models.Silence{
		Start: time.Now().Add(-48 * time.Hour),
		End:   time.Now().Add(-5 * time.Hour),
		Alert: "Foo",
	}

	check(t, sd.AddSilence(silence))
	check(t, sd.AddSilence(past))
	check(t, sd.AddSilence(future))

	active, err := sd.GetActiveSilences()
	check(t, err)
	if len(active) != 1 {
		t.Fatalf("Expected only one active silence. Got %d.", len(active))
	}

	checkIds := func(list map[string]*models.Silence, page int, ids ...string) {
		if len(list) != len(ids) {
			t.Fatalf("Wrong list length. %d != %d", len(list), len(ids))
		}
		for _, id := range ids {
			if list[id] == nil {
				t.Fatalf("Expected to find id %s on page %d, but didn't, %v. %v", id, page, ids, list)
			}
		}
	}

	list, err := sd.ListSilences(2, 0)
	check(t, err)
	checkIds(list, 0, future.ID(), silence.ID())

	list, err = sd.ListSilences(2, 1)
	check(t, err)
	checkIds(list, 1, past.ID())
}
