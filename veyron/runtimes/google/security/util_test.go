package security

import (
	"reflect"
	"testing"

	isecurity "veyron.io/veyron/veyron/runtimes/google/security"
	"veyron.io/veyron/veyron2/security"
)

func mintID() security.PublicID {
	privID, err := isecurity.NewPrivateID("dummy", nil)
	if err != nil {
		panic("couldn't mint new private id")
	}
	return privID.PublicID()
}

func TestSimple(t *testing.T) {
	testCoder(t, mintID())
}

func TestSet(t *testing.T) {
	id := mintID()
	set, err := isecurity.NewSetPublicID(id, id, id)
	if err != nil {
		t.Fatalf("couldn't create set id: %v", err)
	}
	testCoder(t, set)
}

func testCoder(t *testing.T, id security.PublicID) {
	chains, err := EncodeChains(id)
	if err != nil {
		t.Fatalf("couldn't encode public id %v: %v", id.Names(), err)
	}
	decodedID, err := DecodeChains(chains)
	if err != nil {
		t.Fatalf("couldn't decode chains %v: %v", chains, err)
	}
	if expected, actual := id.Names(), decodedID.Names(); !reflect.DeepEqual(expected, actual) {
		t.Fatalf("difference in decoded names, want %v, got %v", expected, actual)
	}
}
