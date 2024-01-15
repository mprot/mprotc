# JavaScript/TypeScript source code translations
The JavaScript and TypeScript source code uses [msgpack-js](https://github.com/mprot/msgpack-js) for the MessagePack encoding.

## Constant
JavaScript and TypeScript:
```js
export const Pi = 3.141592;
```

## Enumeration
JavaScript:
```js
export const E = {
    This: 1,
    That: 2,

    enc: Int.enc,
    dec: Int.dec,
};
```

TypeScript:
```ts
export const E = Int;

// decl.d.ts
export declare const enum E {
    This = 1,
    That = 2,
}
```

## Struct
JavaScript:
```js
export const S = {
    enc(buf, v) { ... },
    dec(buf) { ... },
};
```

TypeScript:
```ts
export const S = {
    enc(buf, v) { ... },
    dec(buf) { ... },
};

// decl.d.ts
export declare var S: Type<S>;

export interface S {
    Foo: number;
    Bar: number;
}
```

## Union
JavaScript:
```js
export const U = {
    enc(buf, v) { ... },
    dec(buf) { ... },
};
```

TypeScript:
```ts
export const U = {
    enc(buf, v) { ... },
    dec(buf) { ... },
};

// decl.d.ts
export declare var U: Type<U>;

export type U = number | S;
```
