// @strict: true
// @noEmit: true

// Skipping an accessor during a check that reports nothing is a deferral, not a waiver. An array of
// schemas is not a Schema, and the constraint has to still be enforced -- from the pass over the
// file's deferred work rather than from the comparison that skipped it.

interface Schema<O> {
    readonly out: O;
}
type Shape = Record<string, Schema<any>>;
declare function object<S extends Shape>(shape: S): Schema<{ [K in keyof S]: S[K]["out"] }> & { shape: S };
declare const leaf: Schema<string>;

const flat = object({
    get bad() {
        return [leaf];
    },
});
export const out = flat.out;
