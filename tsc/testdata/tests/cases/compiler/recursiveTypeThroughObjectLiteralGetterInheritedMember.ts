// @strict: true
// @noEmit: true

// The shape from #62180, where the recursive member is reached through a mapped type that remaps its
// keys by reading each member's own discriminant. Resolving the object schema's members instantiates
// its base with a type that names the schema back, so the schema's table is read while it is still
// being assembled and its inherited `out` is not in it yet.
//
// main reports the schema as missing `out` outright. Answering "absent" in that window instead
// clears that error but resolves the recursive member to `{}`, so every access on it is an error --
// which is why the miss has to be answered "not known yet".

interface Internals<O> {
    optional: "true" | "false";
    out: O;
}

interface StringSchema extends Internals<string> {
    optional: "false";
}

type Shape = Record<string, any>;
type Prettify<T> = { [K in keyof T]: T[K] } & {};
type ObjectOut<S extends Shape> = Prettify<
    {
        [K in keyof S as S[K] extends { optional: "true" } ? K : never]?: S[K]["out"];
    } & {
        [K in keyof S as S[K] extends { optional: "true" } ? never : K]: S[K]["out"];
    }
>;

interface ObjectSchema<S extends Shape> extends Internals<ObjectOut<S>> {
    optional: "false";
}

interface OptionalSchema<T extends Internals<any>> extends Internals<T["out"] | undefined> {
    optional: "true";
}

declare function object<S extends Shape>(shape: S): ObjectSchema<S>;
declare function text(): StringSchema;
declare function optional<T extends Internals<any>>(schema: T): OptionalSchema<T>;

const category = object({
    name: text(),
    get parent() {
        return optional(category);
    },
});

declare const sample: (typeof category)["out"];
// The key that does not recurse resolves fully, and the remapping put the recursive one in the
// optional bucket, so it is `parent?:` rather than `parent:`.
const topName: string = sample.name;
// The recursive one is `any` -- not the type it would be if the table were complete, but usable,
// and the same thing main gives it. This is the line the change is about.
const parentName: string = sample.parent!.name;

// A property that is genuinely absent must still be reported, whoever is asking. The guard only
// applies while a table is mid-assembly, which is a window nothing outside member resolution is in.
interface Base { readonly own: number; }
interface Derived extends Base { readonly extra: string; }
declare const derived: Derived;
// @ts-expect-error
derived.missing;
const ownValue: number = derived.own;
const extraValue: string = derived.extra;
