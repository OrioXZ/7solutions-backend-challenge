# Lottery Search System Design Proposal

The interactive design proposal is available at [`docs/lottery-design-tour.html`](./docs/lottery-design-tour.html).

Because GitHub does not render repository HTML files as live pages, clone or download the repository and open `docs/lottery-design-tour.html` in a browser.

The proposal covers:

- solution architecture and data model;
- wildcard search and indexing strategy;
- randomized allocation without `ORDER BY RANDOM()`;
- atomic reservation using PostgreSQL row locks and `SKIP LOCKED`;
- ticket lifecycle and reservation expiry;
- Big O performance analysis;
- correctness, trade-offs, and production scaling considerations.
