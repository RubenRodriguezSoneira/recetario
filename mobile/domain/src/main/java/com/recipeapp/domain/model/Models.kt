package com.recipeapp.domain.model

import kotlinx.datetime.LocalDateTime

data class Recipe(
    val id: Int,
    val userId: Int,
    val title: String,
    val description: String?,
    val instructions: String,
    val prepTime: Int?,
    val cookTime: Int?,
    val servings: Int,
    val difficulty: Difficulty,
    val imageUrl: String?,
    val isPublic: Boolean,
    val createdAt: LocalDateTime,
    val updatedAt: LocalDateTime,
    val ingredients: List<RecipeIngredient> = emptyList(),
    val tags: List<Tag> = emptyList(),
    val categories: List<Category> = emptyList(),
    val user: User? = null
)

data class RecipeIngredient(
    val id: Int,
    val recipeId: Int,
    val ingredientId: Int,
    val quantity: Double?,
    val unit: String?,
    val notes: String?,
    val ingredient: Ingredient? = null
)

data class Ingredient(
    val id: Int,
    val name: String,
    val category: String?,
    val createdAt: LocalDateTime
)

data class Tag(
    val id: Int,
    val name: String,
    val color: String,
    val createdAt: LocalDateTime
)

data class Category(
    val id: Int,
    val userId: Int,
    val name: String,
    val parentId: Int?,
    val createdAt: LocalDateTime
)

data class User(
    val id: Int,
    val email: String,
    val name: String,
    val isAdmin: Boolean,
    val createdAt: LocalDateTime,
    val updatedAt: LocalDateTime
)

data class RecipeFilter(
    val userId: Int? = null,
    val title: String? = null,
    val difficulty: Difficulty? = null,
    val tags: List<String> = emptyList(),
    val isPublic: Boolean? = null,
    val limit: Int? = null,
    val offset: Int? = null,
    val sortBy: String = "created_at",
    val sortOrder: String = "DESC"
)

enum class Difficulty {
    EASY, MEDIUM, HARD
}

data class AuthResponse(
    val token: String,
    val user: User,
    val expiresIn: Int
)

data class LoginRequest(
    val email: String,
    val password: String
)

data class RegisterRequest(
    val email: String,
    val password: String,
    val name: String
)

data class ApiResponse<T>(
    val message: String,
    val data: T? = null,
    val error: String? = null
)

data class PaginatedResponse<T>(
    val items: List<T>,
    val totalCount: Int,
    val currentPage: Int,
    val pageSize: Int,
    val hasNextPage: Boolean,
    val hasPreviousPage: Boolean
)