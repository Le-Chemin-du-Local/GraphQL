# Le Chemin du Local
[![Gitmoji](https://img.shields.io/badge/gitmoji-%20😜%20😍-FFDD67.svg?style=flat-squar)](https://gitmoji.dev/)
[![Go Reference](https://pkg.go.dev/badge/github.com/99designs/gqlgen.svg)](https://pkg.go.dev/github.com/99designs/gqlgen)

Vous êtes sur le repository de l'API du Chemin du Local. Cette API est basé sur [GraphQL](https://graphql.org/) et [gqlgen](https://github.com/99designs/gqlgen)

## GrapiQL

Ce repository contient un fichier répertoriant l'ensemble des requêtes qui permettent de tester l'API. Vous pourrez trouver ce fichier ici : [GraphiQL.md](https://github.com/Le-Chemin-du-Local/GraphQL/blob/master/GraphiQL.md)

### Exemple de requête 

Ceci est un exemple de requête relativement complexe qui permet de récupérer les commerces et les produits.

```graphql
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
```