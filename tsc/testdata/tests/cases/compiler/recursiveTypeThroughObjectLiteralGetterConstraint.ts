// @strict: true
// @noEmit: true

// Skipping an accessor during a check that reports nothing is a deferral only where the accessor
// genuinely cannot answer. Where it can, skipping it stops the constraint being checked at all and
// the compiler accepts code it should reject: an array of schemas is not a Schema.

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
