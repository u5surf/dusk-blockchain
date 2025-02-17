package user

import (
	"bytes"
	"math/big"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test if removing members from the VotingCommittee works properly.
func TestRemove(t *testing.T) {
	nr := 5
	committee := newCommittee()
	for i := 0; i < nr; i++ {
		k, _ := NewRandKeys()
		bk := (&big.Int{}).SetBytes(k.BLSPubKeyBytes)
		committee.Set = append(committee.Set, bk)
	}
	sort.Sort(committee.Set)

	lastElem := committee.Set[nr-1].Bytes()
	committee.Set.Remove(lastElem)
	i, found := committee.Set.IndexOf(lastElem)
	assert.False(t, found)
	assert.Equal(t, nr-1, i)
}

type sortedKeys []Keys

func (s sortedKeys) Len() int      { return len(s) }
func (s sortedKeys) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s sortedKeys) Less(i, j int) bool {

	return btoi(s[i]).Cmp(btoi(s[j])) < 0
}

func btoi(k Keys) *big.Int {
	b := k.BLSPubKeyBytes
	return (&big.Int{}).SetBytes(b)
}

// Test if MemberKeys returns all public keys in the correct order.
func TestMemberKeys(t *testing.T) {

	p, _, err := NewProvisioners(nil)
	if err != nil {
		t.Fatal(err)
	}

	var ks sortedKeys
	for i := 0; i < 50; i++ {
		keys, _ := NewRandKeys()
		if err := p.AddMember(keys.EdPubKeyBytes, keys.BLSPubKeyBytes, 500, 0, 1000); err != nil {
			t.Fatal(err)
		}
		ks = append(ks, keys)
	}

	sort.Sort(ks)
	v := p.CreateVotingCommittee(1, 500*50, 1, 50)
	mk := v.MemberKeys()
	assert.Equal(t, 50, len(mk))
	for i := 0; i < 3; i++ {
		assert.True(t, bytes.Equal(mk[i], v.Set[i].Bytes()))
	}

}
