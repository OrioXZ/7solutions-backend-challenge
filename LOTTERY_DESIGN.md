# Lottery Search System Design Proposal

The interactive design proposal is available here:

- [Open the Lottery Search System Design Tour](https://orioxz.github.io/7solutions-backend-challenge/lottery-design-tour.html)

The proposal covers:

- solution architecture and data model;
- wildcard search and indexing strategy;
- randomized allocation without `ORDER BY RANDOM()`;
- atomic reservation using PostgreSQL row locks and `SKIP LOCKED`;
- ticket lifecycle and reservation expiry;
- Big O performance analysis;
- correctness, trade-offs, and production scaling considerations.

The source HTML is also available at [`docs/lottery-design-tour.html`](./docs/lottery-design-tour.html).
