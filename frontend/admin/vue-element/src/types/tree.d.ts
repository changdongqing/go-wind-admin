declare global {
  type TreeActionType =
    | "COLLAPSE_ALL"
    | "EXPAND_ALL"
    | "HIERARCHICAL_ASSOCIATION"
    | "HIERARCHICAL_INDEPENDENCE"
    | "SELECT_ALL"
    | "UNSELECT_ALL";
}

export {};
