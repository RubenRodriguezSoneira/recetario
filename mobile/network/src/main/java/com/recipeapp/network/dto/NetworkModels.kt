package com.recipeapp.network.dto

import com.google.gson.annotations.SerializedName

data class RecipeDto(
    @SerializedName("id") val id: Int,
    @SerializedName("user_id") val userId: Int,
    @SerializedName("title") val title: String,
    @SerializedName("description") val description: String?,
    @SerializedName("instructions") val instructions: String,
    @SerializedName("prep_time") val prepTime: Int?,
    @SerializedName("cook_time") val cookTime: Int?,
    @SerializedName("servings") val servings: Int,
    @SerializedName("difficulty") val difficulty: String,
    @SerializedName("image_url") val imageUrl: String?,
    @SerializedName("is_public") val isPublic: Boolean,
    @SerializedName("created_at") val createdAt: String,
    @SerializedName("updated_at") val updatedAt: String,
    @SerializedName("ingredients") val ingredients: List<RecipeIngredientDto> = emptyList(),
    @SerializedName("tags") val tags: List<TagDto> = emptyList(),
    @SerializedName("categories") val categories: List<CategoryDto> = emptyList(),
    @SerializedName("user") val user: UserDto? = null
)

data class RecipeIngredientDto(
    @SerializedName("id") val id: Int,
    @SerializedName("recipe_id") val recipeId: Int,
    @SerializedName("ingredient_id") val ingredientId: Int,
    @SerializedName("quantity") val quantity: Double?,
    @SerializedName("unit") val unit: String?,
    @SerializedName("notes") val notes: String?,
    @SerializedName("ingredient") val ingredient: IngredientDto? = null
)

data class IngredientDto(
    @SerializedName("id") val id: Int,
    @SerializedName("name") val name: String,
    @SerializedName("category") val category: String?,
    @SerializedName("created_at") val createdAt: String
)

data class TagDto(
    @SerializedName("id") val id: Int,
    @SerializedName("name") val name: String,
    @SerializedName("color") val color: String,
    @SerializedName("created_at") val createdAt: String
)

data class CategoryDto(
    @SerializedName("id") val id: Int,
    @SerializedName("user_id") val userId: Int,
    @SerializedName("name") val name: String,
    @SerializedName("parent_id") val parentId: Int?,
    @SerializedName("created_at") val createdAt: String
)

data class UserDto(
    @SerializedName("id") val id: Int,
    @SerializedName("email") val email: String,
    @SerializedName("name") val name: String,
    @SerializedName("is_admin") val isAdmin: Boolean,
    @SerializedName("created_at") val createdAt: String,
    @SerializedName("updated_at") val updatedAt: String
)

data class AuthResponseDto(
    @SerializedName("token") val token: String,
    @SerializedName("user") val user: UserDto,
    @SerializedName("expires_in") val expiresIn: Int
)