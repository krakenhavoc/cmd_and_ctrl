package game

// newTriggeredItemForTest is NewTriggeredItem with a closure Effect:
// a hand-built, UNKEYED trigger item, which only a test may make.
//
// ADR 0041 P9 (#1497, tier 4-final) took NewTriggeredItem's effect
// parameter away, because in production what a stack item does is
// always data — a catalog row the engine names on the item, or a
// registered body. Tests still build items directly to drive the
// engine without a catalog, and they get this helper rather than the
// parameter back. The census counts such an item in
// IntrinsicAbilityCards (snapshotStackItemAs).
func newTriggeredItemForTest(source *Card, label string, effect func(g *Game, item *StackItem) error) *StackItem {
	item := NewTriggeredItem(source, label)
	item.Effect = effect
	return item
}
