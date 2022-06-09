# Processed DIM wishlist

Here you can find DIM wishlist automatically derived from DIM's default
[voltron.txt](https://raw.githubusercontent.com/48klocs/dim-wish-list-sources/master/voltron.txt)
to leave only perks in 3rd and 4th columns.

## How to use this?

DIM allows you to specify multiple wishlist URLs, so you can put this string in
the corresponding field in DIM settings:

 * `https://raw.githubusercontent.com/48klocs/dim-wish-list-sources/master/voltron.txt|https://raw.githubusercontent.com/imax9000/dim-wish-list-sources/generated/voltron.txt`

Note on the ordering: first match wins, so it makes sense to put the default
wishlist first - then if an item matches all 4 perks, you'll see them in the
UI. And if there's no match for all 4 columns, it might still match 3rd and
4th.

## Variants

You can see all available options in the `generated` branch in this repo.

### `voltron.txt`

This one has all the notes preserved as they are in the default wishlist. As a
consequence, on the
[Armory screen](https://destinyitemmanager.fandom.com/wiki/Armory) DIM will
merge in all
entries with corresponding entries from the default wishlist and you'll see
only combinations will all 4 perks.

Use it by putting the following string in the DIM settings: `https://raw.githubusercontent.com/48klocs/dim-wish-list-sources/master/voltron.txt|https://raw.githubusercontent.com/imax9000/dim-wish-list-sources/generated/voltron.txt`

### `voltron_flagged.txt`

This one has `[3rd&4th columns only]` added to every note, so on the
[Armory screen](https://destinyitemmanager.fandom.com/wiki/Armory) DIM will
group them separately.

Use it by putting the following string in the DIM settings: `https://raw.githubusercontent.com/48klocs/dim-wish-list-sources/master/voltron.txt|https://raw.githubusercontent.com/imax9000/dim-wish-list-sources/generated/voltron_flagged.txt`
