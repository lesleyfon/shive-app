# Shive API

Shive API is a movie management system built with Go (Gin framework) and MongoDB. It provides functionality for managing users, movies, genres, and reviews with authentication and authorization.

## Features

- User Authentication & Authorization
- Movie Management
- Genre Management
- Review System
- Role-based Access Control (Admin/User)
- MongoDB Integration
- JWT Token Authentication
- RESTful API Design

## Tech Stack

- **Go** - Programming Language
- **Gin** - Web Framework
- **MongoDB** - Database
- **JWT** - Authentication
- **GOW** - [Live Reload for Development](https://github.com/mitranim/gow)
- **Gitingest** - [Generating App Structure](https://gitingest.com/)

## Prerequisites

- Go 1.23.2 or higher
- MongoDB
- Git

## Environment Variables

Create a `.env` file in the root directory with the following variables:

```env
PORT=9000
MONGOURI=your_mongodb_connection_string
ENV=development
SECRET_KEY=your_jwt_secret_key

```

## Installation & Setup

1. Clone the repository:

```bash
git clone https://github.com/lesleyfon/shive-app.git
cd shive-app
```

2. Install dependencies:

```bash
go mod download
```

### Development with live reload

```bash
gow run ./      
```

### Production

```bash
go run main.go
```



## API Endpoints

### Authentication
- `POST /users/login` - User login
- `POST /users/signup` - User registration

### Users
- `GET /users` - Get all users (Admin only)
- `GET /users/:user_id` - Get user by ID
- `PUT /users/:user_id` - Update user
- `DELETE /users/:user_id` - Delete user

### Movies
- `POST /movies/create-movie` - Create new movie (Admin only)
- `GET /movies` - Get all movies
- `GET /movies/:movie_id` - Get movie by ID
- `PUT /movies/:movie_id` - Update movie (Admin only)
- `DELETE /movies/:movie_id` - Delete movie (Admin only)
- `GET /movies/search/:name` - Search movies by name
- `GET /movies/filter/:genre_id` - Filter movies by genre

### Genres
- `POST /genres/creategenre` - Create new genre (Admin only)
- `GET /genres` - Get all genres
- `GET /genres/:genre_id` - Get genre by ID
- `PUT /genres/:genre_id` - Update genre (Admin only)
- `DELETE /genres/:genre_id` - Delete genre (Admin only)
- `GET /genres/search-genre` - Search genres by name

### Reviews
- `POST /review/add-review` - Add new review (User only)
- `GET /review/filter/:movie_id` - Get reviews by movie ID
- `DELETE /review/delete/:review_id` - Delete review

## Authentication

The API uses JWT tokens for authentication. Include the token in the request header:
```
token: <your_token>
```


## Role-Based Access

- **Admin**: Can perform all operations
- **User**: Can view content and manage reviews

## Testing

Run the test suite:

```bash
go test -v ./test
```


## Project Structure

```
Directory structure:
└──shive-app/
    ├── api-test.shell # API test
    ├── go.mod                  # Go module
    ├── go.sum                  # Go sum
    ├── main.go                 # Main file
    ├── shive-app               # Shive app
    ├── .air.toml               # Air config
    ├── controllers/
    │   ├── genreController.go  # Genre controller
    │   ├── movieController.go  # Movie controller
    │   ├── reviewController.go # Review controller
    │   └── userController.go   # User controller
    ├── database/
    │   └── db.go               # Database config
    ├── helpers/
    │   ├── authHelper.go       # Auth helper
    │   └── tokenHelper.go      # Token helper
    ├── middleware/
    │   └── authMiddleware.go   # Auth middleware
    ├── models/
    │   ├── genreModel.go       # Genre model
    │   ├── movieModel.go       # Movie model
    │   ├── reviewModel.go      # Review model
    │   └── userModel.go        # User model
    ├── routes/
    │   ├── authRouter.go       # Auth router
    │   ├── genreRouter.go      # Genre router
    │   ├── movieRouter.go      # Movie router
    │   ├── reviewRouter.go     # Review router
    │   └── userRouter.go       # User router
    ├── test/
    │   └── auth_test.go        # Auth test
    └── .github/
        └── workflows/
            └── go.yml          # Github actions

```


## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feat/new-feature`)
3. Commit your changes (`git commit -m 'Add some new feature'`)
4. Push to the branch (`git push origin feat/new-feature`)
5. Open a Pull Request

## License
This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For support, email lesleybanadzem2017@gmail.com or open an issue in the repository.
