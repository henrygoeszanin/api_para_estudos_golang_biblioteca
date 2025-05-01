package services

import (
	"errors"

	"github.com/henrygoeszanin/api_golang_estudos/application/dtos"
	"github.com/henrygoeszanin/api_golang_estudos/application/interfaces/repositories"
	"github.com/henrygoeszanin/api_golang_estudos/domain/entities"
)

// BookService implementa a interface BookService
type BookService struct {
	bookRepository repositories.BookRepository
}

// NewBookService cria uma nova instância do serviço de livros
func NewBookService(bookRepository repositories.BookRepository) *BookService {
	return &BookService{
		bookRepository: bookRepository,
	}
}

// Create cria um novo livro
func (bookservice *BookService) Create(bookDTO dtos.BookCreateDTO) (*dtos.BookResponseDTO, error) {
	book := entities.Book{
		Title:       bookDTO.Title,
		Author:      bookDTO.Author,
		Description: bookDTO.Description,
		Quantity:    bookDTO.Quantity,
		Available:   bookDTO.Quantity, // Inicialmente todos disponíveis
	}

	if err := bookservice.bookRepository.Create(&book); err != nil {
		return nil, err
	}

	responseDTO := dtos.BookToResponseDTO(book)
	return &responseDTO, nil
}

// GetByID busca um livro pelo ID
func (bookservice *BookService) GetByID(id uint) (*dtos.BookResponseDTO, error) {
	book, err := bookservice.bookRepository.FindByID(id)
	if err != nil {
		return nil, err
	}
	if book == nil {
		return nil, errors.New("livro não encontrado")
	}

	responseDTO := dtos.BookToResponseDTO(*book)
	return &responseDTO, nil
}

// List retorna todos os livros
func (bookservice *BookService) List() ([]dtos.BookResponseDTO, error) {
	books, err := bookservice.bookRepository.List()
	if err != nil {
		return nil, err
	}

	var bookDTOs []dtos.BookResponseDTO
	for _, book := range books {
		bookDTOs = append(bookDTOs, dtos.BookToResponseDTO(*book))
	}

	return bookDTOs, nil
}

// Update atualiza os dados de um livro
func (bookservice *BookService) Update(id uint, bookDTO dtos.BookUpdateDTO) (*dtos.BookResponseDTO, error) {
	book, err := bookservice.bookRepository.FindByID(id)
	if err != nil {
		return nil, err
	}
	if book == nil {
		return nil, errors.New("livro não encontrado")
	}

	// Atualizar campos se fornecidos
	if bookDTO.Title != "" {
		book.Title = bookDTO.Title
	}
	if bookDTO.Author != "" {
		book.Author = bookDTO.Author
	}
	if bookDTO.Description != "" {
		book.Description = bookDTO.Description
	}
	if bookDTO.Quantity > 0 {
		// Atualizar também o disponível proporcionalmente
		diff := bookDTO.Quantity - book.Quantity
		book.Available = book.Available + diff
		if book.Available < 0 {
			book.Available = 0
		}
		book.Quantity = bookDTO.Quantity
	}

	if err := bookservice.bookRepository.Update(book); err != nil {
		return nil, err
	}

	responseDTO := dtos.BookToResponseDTO(*book)
	return &responseDTO, nil
}

// Delete remove um livro
func (bookservice *BookService) Delete(id uint) error {
	return bookservice.bookRepository.Delete(id)
}
