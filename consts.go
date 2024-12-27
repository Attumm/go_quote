package main

const SwaggerSpec = `
{
  "openapi": "3.0.0",
  "info": {
    "title": "Quotes API",
    "version": "1.0.0",
    "description": "An API for quotes, authors, and tags."
  },
  "paths": {
   "/quotes": {
      "get": {
        "summary": "List all quotes",
        "responses": {
          "200": {
            "description": "Successful response",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/PaginatedQuotes"
                }
              }
            }
          }
        },
        "parameters": [
          {
            "$ref": "#/components/parameters/PageParam"
          },
          {
            "$ref": "#/components/parameters/PageSizeParam"
          },
          {
            "$ref": "#/components/parameters/FormatParam"
          }
        ]
      }
    },
    "/random-quote": {
      "get": {
        "summary": "Get a random quote",
        "responses": {
          "200": {
            "description": "Successful response",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/ResponseQuote"
                }
              }
            }
          },
          "404": {
            "description": "No quotes available"
          }
        },
        "parameters": [
          {
            "$ref": "#/components/parameters/FormatParam"
          }
        ]
      }
    },
    "/quote/{quoteId}": {
      "get": {
        "summary": "Get a specific quote by ID",
        "parameters": [
          {
            "name": "quoteId",
            "in": "path",
            "required": true,
            "schema": {
              "type": "integer"
            }
          },
          {
            "$ref": "#/components/parameters/FormatParam"
          }
        ],
        "responses": {
          "200": {
            "description": "Successful response",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/ResponseQuote"
                }
              }
            }
          },
          "404": {
            "description": "Quote not found"
          }
        }
      }
    },
    "/tags": {
      "get": {
        "summary": "List all tags",
        "responses": {
          "200": {
            "description": "Successful response",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/PaginatedTags"
                }
              }
            }
          }
        },
        "parameters": [
          {
            "$ref": "#/components/parameters/PageParam"
          },
          {
            "$ref": "#/components/parameters/PageSizeParam"
          },
          {
            "$ref": "#/components/parameters/FormatParam"
          }
        ]
      }
    },
    "/tags/{tagId}": {
      "get": {
        "summary": "Get quotes for a specific tag",
        "parameters": [
          {
            "name": "tagId",
            "in": "path",
            "required": true,
            "schema": {
              "type": "string"
            }
          },
          {
            "$ref": "#/components/parameters/PageParam"
          },
          {
            "$ref": "#/components/parameters/PageSizeParam"
          },
          {
            "$ref": "#/components/parameters/FormatParam"
          }
        ],
        "responses": {
          "200": {
            "description": "Successful response",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/PaginatedQuotes"
                }
              }
            }
          },
          "404": {
            "description": "Tag not found"
          }
        }
      }
    },
    "/authors": {
      "get": {
        "summary": "List all authors",
        "responses": {
          "200": {
            "description": "Successful response",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/PaginatedAuthors"
                }
              }
            }
          }
        },
        "parameters": [
          {
            "$ref": "#/components/parameters/PageParam"
          },
          {
            "$ref": "#/components/parameters/PageSizeParam"
          },
          {
            "$ref": "#/components/parameters/FormatParam"
          }
        ]
      }
    },
    "/authors/{authorId}": {
      "get": {
        "summary": "Get quotes for a specific author",
        "parameters": [
          {
            "name": "authorId",
            "in": "path",
            "required": true,
            "schema": {
              "type": "string"
            }
          },
          {
            "$ref": "#/components/parameters/PageParam"
          },
          {
            "$ref": "#/components/parameters/PageSizeParam"
          },
          {
            "$ref": "#/components/parameters/FormatParam"
          }
        ],
        "responses": {
          "200": {
            "description": "Successful response",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/AuthorQuotes"
                }
              }
            }
          },
          "404": {
            "description": "Author not found"
          }
        }
      }
    }
  },
  "components": {
    "schemas": {
      "Quote": {
        "type": "object",
        "properties": {
          "text": {
            "type": "string"
          },
          "author": {
            "type": "string"
          },
          "tags": {
            "type": "array",
            "items": {
              "type": "string"
            }
          }
        }
      },
      "ResponseQuote": {
        "allOf": [
          {
            "$ref": "#/components/schemas/Quote"
          },
          {
            "type": "object",
            "properties": {
              "id": {
                "type": "integer"
              },
              "authorId": {
                "type": "string"
              }
            }
          }
        ]
      },
      "Pagination": {
        "type": "object",
        "properties": {
          "page": {
            "type": "integer"
          },
          "pageSize": {
            "type": "integer"
          },
          "totalPages": {
            "type": "integer"
          },
          "totalItems": {
            "type": "integer"
          }
        }
      },
      "PaginatedQuotes": {
        "type": "object",
        "properties": {
          "quotes": {
            "type": "array",
            "items": {
              "$ref": "#/components/schemas/ResponseQuote"
            }
          },
          "pagination": {
            "$ref": "#/components/schemas/Pagination"
          }
        }
      },
      "AuthorResponse": {
        "type": "object",
        "properties": {
          "name": {
            "type": "string"
          },
          "authorId": {
            "type": "string"
          },
          "totalQuotes": {
            "type": "integer"
          }
        }
      },
      "PaginatedAuthors": {
        "type": "object",
        "properties": {
          "authors": {
            "type": "array",
            "items": {
              "$ref": "#/components/schemas/AuthorResponse"
            }
          },
          "pagination": {
            "$ref": "#/components/schemas/Pagination"
          }
        }
      },
      "TagResponse": {
        "type": "object",
        "properties": {
          "name": {
            "type": "string"
          },
          "tagId": {
            "type": "string"
          },
          "totalQuotes": {
            "type": "integer"
          }
        }
      },
      "PaginatedTags": {
        "type": "object",
        "properties": {
          "tags": {
            "type": "array",
            "items": {
              "$ref": "#/components/schemas/TagResponse"
            }
          },
          "pagination": {
            "$ref": "#/components/schemas/Pagination"
          }
        }
      },
      "AuthorQuotes": {
        "type": "object",
        "properties": {
          "author": {
            "type": "string"
          },
          "authorId": {
            "type": "string"
          },
          "totalQuotes": {
            "type": "integer"
          },
          "quotes": {
            "type": "array",
            "items": {
              "$ref": "#/components/schemas/ResponseQuote"
            }
          },
          "pagination": {
            "$ref": "#/components/schemas/Pagination"
          }
        }
      }
    },
    "parameters": {
      "PageParam": {
        "name": "page",
        "in": "query",
        "schema": {
          "type": "integer",
          "default": 1
        },
        "description": "Page number for pagination"
      },
      "PageSizeParam": {
        "name": "page_size",
        "in": "query",
        "schema": {
          "type": "integer",
          "default": 20
        },
        "description": "Number of items per page"
      },
      "FormatParam": {
        "name": "format",
        "in": "query",
        "schema": {
          "type": "string",
          "enum": ["xml", "html", "json", "text", "markdown", "yaml", "csv", "rss", "atom", "oembed", "oembed.xml", "embed", "embed.js", "svg", "svg-download", "wav"],
          "default": "json"
        },
        "description": "Output format of the response"
      }
    }
  }
}
`
