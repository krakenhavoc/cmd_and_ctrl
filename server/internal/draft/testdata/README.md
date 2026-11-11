# draft testdata

`botset_dump.json` is the draft bot's fixture set, `c2d` ("C2 Fixture
Draft"): a small file in the shape of Scryfall's default-cards dump, **made
up for the tests**. It is what the bot is tested and measured on in CI,
where there is no Scryfall dump (ADR 0222 §7, #5439). It is not Scryfall
data and says nothing about any real set.

What is in it (114 cards: 73 commons, 27 uncommons, 10 rares, 4 mythics):

- **79 real cards the catalog holds as Full**, with their real oracle IDs
  and names. Their oracle text is copied from the catalog's own oracle
  fixtures (`internal/cards/coverage/testdata/oracle/<oracle_id>.json`).
  Their mana cost, type line, power, toughness and keywords were typed by
  hand and are the cards' printed ones. **The set, the rarities, the
  collector numbers and the printing IDs are invented.**
- **30 invented vanilla creatures**, six commons a colour (`C2 Fixture
  White Cub` and so on: a 2/2 for two up to a 6/5 for six), with invented
  oracle IDs. They have no catalog entry and need none: the engine plays a
  creature with no abilities as it is.
- **The five basic lands**, with invented oracle IDs, as in the draftpacks
  fixture.

`loadBotFixture` (`bot_fixture_test.go`) fails if a real card of the set
stops being Full in the catalog, and `TestTheFixtureSetIsWhatItsReadmeSays`
pins the counts above.

A change to this file moves every number the measure reports
(`bot_measure_test.go`, and `internal/botarena/draft_decks_test.go` for the
games). Change it on purpose only, and report the measure again.
