# GraphiQL

Ce fichier contient les requêtes créés pour tester l'API. Vous pouvez les copier/coller lorsque vous lancez l'API.

```graphql
# Welcome to GraphiQL
#
# GraphiQL is an in-browser tool for writing, validating, and
# testing GraphQL queries.
#
# Type queries into this side of the screen, and you will see intelligent
# typeaheads aware of the current GraphQL type schema and live syntax and
# validation errors highlighted within the text.
#
# GraphQL queries typically start with a "{" character. Lines that start
# with a # are ignored.
#
# An example GraphQL query might look like:
#
#     {
#       field(arg: "value") {
#         subField
#       }
#     }
#
# Keyboard shortcuts:
#
#  Prettify Query:  Shift-Ctrl-P (or press the prettify button above)
#
#     Merge Query:  Shift-Ctrl-M (or press the merge button above)
#
#       Run Query:  Ctrl-Enter (or press the play button above)
#
#   Auto Complete:  Ctrl-Space (or just start typing)
#

##################
## UTILISATEURS ##
##################

# ----- #
# QUERY #
# ----- #

query getUsers {
  users {
    email,
    firstName,
    commerce {
      name,
      description
    }
  }
}

query getUser {
  user(id: "620b66e75fd472177448f253") {
    email,
    firstName,
    commerce {
      name,
      description
    }
  }
}

# -------- #
# MUTATION #
# -------- #

# INSCRIPTION

mutation createUser1 {
  createUser(input: {
    email: "admin@me.com",
    password: "dE8bdTUE"
  }) {
    id,
    email,
    firstName
  }
}


mutation createUser2 {
  createUser(input: {
    email: "commercant@me.com",
    password: "dE8bdTUE",
    firstName: "Roger"
  }) {
    id,
    email,
    firstName
  }
}

# CONNEXION

# Token : eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2NDUwMDA5ODQsImlkIjoiNjIwYjY2ZTc1ZmQ0NzIxNzc0NDhmMjUzIn0.E8KybuzQDlWRroqXK93dwVw_dTK1d1kk4_qlpiXC8sk
mutation login1 {
  login(input: {
    email: "admin@me.com",
    password: "dE8bdTUE"
  })
}

# Token : eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2NDUwMDEwMjQsImlkIjoiNjIwYjY2ZjE1ZmQ0NzIxNzc0NDhmMjU0In0.uwGpn_lr1OVOOu74Gg7du7qlTDMz9JUP6iqPF2Exypk
mutation login2 {
  login(input: {
    email: "commercant@me.com",
    password: "dE8bdTUE"
  })
}

mutation failedLogin1 {
  login(input: {
    email: "admin@me.com",
    password: "dE8bdUE"
  })
}

mutation failedLogin2 {
  login(input: {
    email: "adm@me.com",
    password: "dE8bdTUE"
  })
}

###############
## COMMERCES ##
###############

# ----- #
# QUERY #
# ----- #

query getCommerces {
  commerces {
    edges {
      node {
        name,
        description,
        storekeeper {
          firstName
        }
      }
    },
    pageInfo {
      hasNextPage
    }
  }
}

query getCommerce {
  commerce(id: "620b6d99df37729e9e3dd7ea") {
    name,
    description
    storekeeper {
      firstName
    },
  }
}

# -------- #
# MUTATION #
# -------- #

mutation createCommerce {
  createCommerce(input: {
    name: "La Bizhhh",
    description: "Une brasserie situé près du Rheu",
    address: "54 Rue Nationale, 35650 Le Rheu"
    storekeeperWord: "Venez expérimenter la bière comme jamais avant",
    phone: "0685811697",
    email: "contact@bizhhh.bzh"
    
  }) {
    storekeeper {
      firstName,
      lastName,
    },
    name
  }
}

##############
## PRODUITS ##
##############

# ----- #
# QUERY #
# ----- #

# Les produits ne peuvent être accédé que par un commerce

query getAllProducts {
  commerces {
    edges {
      node {
        name,
        description,
        storekeeper {
          firstName
        },
        procuts(first: 1) {
          edges {
            node {
              name,
              description,
              price
            }
          },
          pageInfo {
            startCursor,
            endCursor
            hasNextPage
          }
        },
      }
    },
    pageInfo {
      hasNextPage
    }
  }
}

query getProductBizhhh {
  commerce(id: "620b6d99df37729e9e3dd7ea") {
    name,
    description
    storekeeper {
      firstName
    },
    procuts {
      edges {
        node {
          name,
          description,
          price
        }
      },
      pageInfo {
        startCursor,
        endCursor
        hasNextPage
      }
    },
  }
}

# -------- #
# MUTATION #
# -------- #

mutation createProductFromAdmin {
  createProduct(input: {
    commerceID: "620b6d99df37729e9e3dd7ea",
    name: "Bière blonde",
    description: "Une supère bière blonde !",
    price: 2,
    unit: "unité",
    isBreton: true,
    categories: "Bières & cidres"
  }) {
    name,
    description
  }
}

mutation createProductFromStoreKeeper {
  createProduct(input: {
    name: "Bière ambrée",
    description: "Une supère bière ambrée !",
    price: 2,
    unit: "unité",
    isBreton: true,
    categories: "Bières & cidres"
  }) {
    name,
    description
  }
}
```