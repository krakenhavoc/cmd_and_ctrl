package game

import (
	"reflect"
	"testing"

	"github.com/google/uuid"
)

// Every CardDef slot says how it bears on the #2884 independence check.
// A new slot fails here until it is classified: an unclassified slot
// already counts as rules-bearing (the board is opaque), so forgetting
// it only costs prompts, but the class should be a decision.
func TestEveryCardDefSlotHasAnIndependenceClass(t *testing.T) {
	ty := reflect.TypeOf(CardDef{})
	seen := map[string]bool{}
	for i := 0; i < ty.NumField(); i++ {
		f := ty.Field(i)
		if !f.IsExported() {
			continue
		}
		seen[f.Name] = true
		if _, ok := cardDefFieldClass[f.Name]; !ok {
			t.Errorf("CardDef.%s has no independence class in cardDefFieldClass (trigger_independence.go)", f.Name)
		}
	}
	for name := range cardDefFieldClass {
		if !seen[name] {
			t.Errorf("cardDefFieldClass names %q, which is not a CardDef slot", name)
		}
	}
}

func TestDefBearsRules(t *testing.T) {
	if g, o := defBearsRules(&CardDef{Activated: []ActivatedAbilityShape{{}}}); g || o {
		t.Errorf("an activated ability bears no rules: %v %v", g, o)
	}
	if g, o := defBearsRules(&CardDef{Static: []StaticAbility{{}}}); g || !o {
		t.Errorf("a static ability is its own permanent's rules: %v %v", g, o)
	}
	if g, _ := defBearsRules(&CardDef{CantGainLife: []CantGainLifeStatic{{}}}); !g {
		t.Error("a can't-gain-life slot makes the board opaque")
	}
}

func TestMentionsAnyIDFindsAnIDAnywhere(t *testing.T) {
	id := uuid.New()
	ids := map[uuid.UUID]bool{id: true}
	type inner struct {
		ref ObjectRef
	}
	type outer struct {
		Name  string
		Inner *inner
		List  []TargetRef
		Map   map[string]uuid.UUID
		Fn    func()
	}
	for _, c := range []struct {
		name string
		v    any
		want bool
	}{
		{"an unexported nested ref", outer{Inner: &inner{ref: ObjectRef{ID: id}}}, true},
		{"a slice element", outer{List: []TargetRef{{Kind: TargetCard, ID: id}}}, true},
		{"a map value", outer{Map: map[string]uuid.UUID{"x": id}}, true},
		{"nothing", outer{Name: "x", List: []TargetRef{{ID: uuid.New()}}, Fn: func() {}}, false},
	} {
		if got := mentionsAnyID(reflect.ValueOf(c.v), ids, map[uintptr]bool{}); got != c.want {
			t.Errorf("%s: %v, want %v", c.name, got, c.want)
		}
	}
}

// A seat whose items have no declared footprint keeps asking: test-stub
// triggers carry no catalog stamp, so this is the old behaviour, and the
// pure skips still answer first.
func TestUnstampedTriggersKeepAsking(t *testing.T) {
	g := NewGame()
	a := &StackItem{Kind: StackItemTriggered, SourceCardID: uuid.New(), Label: "a"}
	b := &StackItem{Kind: StackItemTriggered, SourceCardID: uuid.New(), Label: "b"}
	if !g.seatNeedsTriggerOrderLocked([]*StackItem{a, b}, TriggerOrderWhenItMatters) {
		t.Error("two undescribed triggers must ask")
	}
	if g.seatNeedsTriggerOrderLocked([]*StackItem{a, b}, TriggerOrderNever) {
		t.Error("never does not ask")
	}
}
