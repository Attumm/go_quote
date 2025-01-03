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
              },
              "application/xml": {
                "schema": {
                  "$ref": "#/components/schemas/ResponseQuote"
                }
              },
              "text/html": {
                "schema": {
                  "type": "string"
                }
              },
              "text/plain": {
                "schema": {
                  "type": "string"
                }
              },
              "text/markdown": {
                "schema": {
                  "type": "string"
                }
              },
              "application/yaml": {
                "schema": {
                  "$ref": "#/components/schemas/ResponseQuote"
                }
              }
            }
          }
        },
        "parameters": [
          {
            "$ref": "#/components/parameters/FormatParam"
          },
          {
            "$ref": "#/components/parameters/AcceptHeader"
          }
        ]
      }
    },
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
    "/quotes/{quoteId}": {
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
          },
          {
            "$ref": "#/components/parameters/AcceptHeader"
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
              },
              "application/xml": {
                "schema": {
                  "$ref": "#/components/schemas/ResponseQuote"
                }
              },
              "text/html": {
                "schema": {
                  "type": "string"
                }
              },
              "text/plain": {
                "schema": {
                  "type": "string"
                }
              },
              "text/markdown": {
                "schema": {
                  "type": "string"
                }
              },
              "application/yaml": {
                "schema": {
                  "$ref": "#/components/schemas/ResponseQuote"
                }
              }
            }
          },
          "404": {
            "description": "Quote not found",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/ErrorResponse"
                }
              },
              "application/xml": {
                "schema": {
                  "$ref": "#/components/schemas/ErrorResponse"
                }
              },
              "text/html": {
                "schema": {
                  "type": "string"
                }
              },
              "text/plain": {
                "schema": {
                  "type": "string"
                }
              },
              "text/markdown": {
                "schema": {
                  "type": "string"
                }
              },
              "application/yaml": {
                "schema": {
                  "$ref": "#/components/schemas/ErrorResponse"
                }
              }
            }
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
            "description": "Tag not found",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/ErrorResponse"
                }
              }
            }
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
            "description": "Author not found",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/ErrorResponse"
                }
              }
            }
          }
        }
      }
    }
  },
  "components": {
    "responses": {
      "GenericSuccessResponse": {
        "description": "Successful response",
        "content": {
          "application/json": {
            "schema": {
              "$ref": "#/components/schemas/GenericResponse"
            }
          },
          "application/xml": {
            "schema": {
              "$ref": "#/components/schemas/GenericResponse"
            }
          },
          "text/html": {
            "schema": {
              "type": "string"
            }
          },
          "text/plain": {
            "schema": {
              "type": "string"
            }
          },
          "text/markdown": {
            "schema": {
              "type": "string"
            }
          },
          "application/yaml": {
            "schema": {
              "$ref": "#/components/schemas/GenericResponse"
            }
          }
        }
      },
      "GenericErrorResponse": {
        "description": "Error response",
        "content": {
          "application/json": {
            "schema": {
              "$ref": "#/components/schemas/ErrorResponse"
            }
          },
          "application/xml": {
            "schema": {
              "$ref": "#/components/schemas/ErrorResponse"
            }
          },
          "text/html": {
            "schema": {
              "type": "string"
            }
          },
          "text/plain": {
            "schema": {
              "type": "string"
            }
          },
          "text/markdown": {
            "schema": {
              "type": "string"
            }
          },
          "application/yaml": {
            "schema": {
              "$ref": "#/components/schemas/ErrorResponse"
            }
          }
        }
      }
    },
    "schemas": {
      "GenericResponse": {
        "type": "object",
        "description": "A generic response object that can represent any successful response"
      },
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
              "author_id": {
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
          "author_id": {
            "type": "string"
          },
          "total_quotes": {
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
          "tag_id": {
            "type": "string"
          },
          "total_quotes": {
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
          "author_id": {
            "type": "string"
          },
          "total_quotes": {
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
      },
      "ErrorResponse": {
        "type": "object",
        "properties": {
          "status": {
            "type": "integer"
          },
          "message": {
            "type": "string"
          },
          "error": {
            "type": "string"
          }
        },
        "required": ["status", "message", "error"]
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
          "enum": [
			"atom",
			"csv",
			"embed",
			"embed.js",
			"html",
			"json",
			"markdown",
			"oembed",
			"oembed.xml",
			"rss",
			"svg",
			"svg-download",
			"text",
			"wav",
			"xml",
			"yaml"
          ],
          "default": "json"
        },
        "description": "Output format of the response"
      },
      "AcceptHeader": {
        "name": "Accept",
        "in": "header",
        "schema": {
          "type": "string"
        },
        "description": "Accepted content types"
      }
    }
  }
}
`
