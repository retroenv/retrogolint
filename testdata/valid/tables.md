# Valid tables

The separator row keeps the same cell count as the header.

| Stage | Dependency | Deliverable |
| --- | --- | --- |
| 00 | None | Verified baseline |

Alignment markers are valid separator cells.

| Result | Count | Notes |
| :----- | ----: | :---: |
| pass | 1 | first |

An escaped pipe stays inside one cell.

| Pattern | Meaning |
| --- | --- |
| `a \| b` | escaped pipe |

A pipe block inside a code fence is not a table.

```markdown
| Stage | Dependency | Deliverable |
| --- | --- |
| 00 | None | Verified baseline |
```
