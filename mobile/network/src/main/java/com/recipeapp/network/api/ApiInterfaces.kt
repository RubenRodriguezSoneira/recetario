package com.recipeapp.network.api

import com.recipeapp.domain.model.*

interface RecipeApi {
    suspend fun getRecipes(filter: RecipeFilter): List<Recipe>
    suspend fun getRecipe(id: Int): Recipe
    suspend fun createRecipe(recipe: Recipe): Recipe
    suspend fun updateRecipe(id: Int, recipe: Recipe): Recipe
    suspend fun deleteRecipe(id: Int): Unit
}

interface AuthApi {
    suspend fun login(request: LoginRequest): AuthResponse
    suspend fun register(request: RegisterRequest): AuthResponse
    suspend fun refreshToken(): AuthResponse
    suspend fun logout(): Unit
}

interface IngredientApi {
    suspend fun getIngredients(): List<Ingredient>
    suspend fun getIngredient(id: Int): Ingredient
    suspend fun createIngredient(name: String, category: String?): Ingredient
    suspend fun updateIngredient(id: Int, name: String, category: String?): Ingredient
    suspend fun deleteIngredient(id: Int): Unit
    suspend fun searchIngredients(query: String): List<Ingredient>
}

interface TagApi {
    suspend fun getTags(): List<Tag>
    suspend fun getTag(id: Int): Tag
    suspend fun createTag(name: String, color: String?): Tag
    suspend fun updateTag(id: Int, name: String, color: String?): Tag
    suspend fun deleteTag(id: Int): Unit
}