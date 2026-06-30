// RecipeApp JavaScript utilities

function showLoginModal() {
    document.getElementById('loginModal').style.display = 'flex';
}

function hideLoginModal() {
    document.getElementById('loginModal').style.display = 'none';
}

function showRegisterModal() {
    document.getElementById('registerModal').style.display = 'flex';
}

function hideRegisterModal() {
    document.getElementById('registerModal').style.display = 'none';
}

let ingredientCount = 0;
let instructionCount = 0;

function escapeHTML(value) {
    return String(value)
        .replaceAll('&', '&amp;')
        .replaceAll('<', '&lt;')
        .replaceAll('>', '&gt;')
        .replaceAll('"', '&quot;')
        .replaceAll("'", '&#39;');
}

function ingredientRowHTML(index, ingredient) {
    const name = escapeHTML(ingredient?.name || '');
    const amount = escapeHTML(ingredient?.amount || '');
    const unit = escapeHTML(ingredient?.unit || '');
    return `
        <input type="text" name="ingredients[${index}].name" value="${name}" placeholder="ej. Arroz" required class="form-input">
        <input type="text" name="ingredients[${index}].amount" value="${amount}" placeholder="400" required class="form-input">
        <input type="text" name="ingredients[${index}].unit" value="${unit}" placeholder="g" required class="form-input">
        <button type="button" onclick="removeIngredient(this)" class="btn-remove">✕</button>
    `;
}

function instructionRowHTML(index, instruction) {
    const text = escapeHTML(instruction?.text || '');
    return `
        <span class="step-label">${index + 1}.</span>
        <textarea name="instructions[${index}].text" placeholder="Paso ${index + 1}" rows="2" required class="form-textarea" style="flex:1;">${text}</textarea>
        <button type="button" onclick="removeInstruction(this)" class="btn-remove" style="align-self:flex-start; margin-top:0.4rem;">✕</button>
    `;
}

function addIngredient(ingredient) {
    const ingredientsList = document.getElementById('ingredients-list');
    if (!ingredientsList) return;

    const newIngredient = document.createElement('div');
    newIngredient.className = 'ingredient-row';
    newIngredient.innerHTML = ingredientRowHTML(ingredientCount, ingredient);
    ingredientsList.appendChild(newIngredient);
    ingredientCount += 1;
}

function removeIngredient(button) {
    const ingredientItem = button.closest('.ingredient-row');
    if (ingredientItem) {
        ingredientItem.remove();
        renumberIngredientRows();
    }
}

function renumberIngredientRows() {
    const rows = document.querySelectorAll('#ingredients-list .ingredient-row');
    rows.forEach((row, index) => {
        const nameInput = row.querySelector('input[name*=".name"]');
        const amountInput = row.querySelector('input[name*=".amount"]');
        const unitInput = row.querySelector('input[name*=".unit"]');
        if (nameInput) nameInput.name = `ingredients[${index}].name`;
        if (amountInput) amountInput.name = `ingredients[${index}].amount`;
        if (unitInput) unitInput.name = `ingredients[${index}].unit`;
    });
    ingredientCount = rows.length;
}

function addInstruction(instruction) {
    const instructionsList = document.getElementById('instructions-list');
    if (!instructionsList) return;

    const newInstruction = document.createElement('div');
    newInstruction.className = 'instruction-row';
    newInstruction.innerHTML = instructionRowHTML(instructionCount, instruction);
    instructionsList.appendChild(newInstruction);
    instructionCount += 1;
}

function removeInstruction(button) {
    const instructionItem = button.closest('.instruction-row');
    if (instructionItem) {
        instructionItem.remove();
        renumberInstructionRows();
    }
}

function renumberInstructionRows() {
    const instructions = document.querySelectorAll('#instructions-list .instruction-row');

    instructions.forEach((item, index) => {
        const label = item.querySelector('.step-label');
        const textarea = item.querySelector('textarea');

        if (label) {
            label.textContent = `${index + 1}.`;
        }

        if (textarea) {
            textarea.name = `instructions[${index}].text`;
            textarea.placeholder = `Paso ${index + 1}`;
        }
    });

    instructionCount = instructions.length;
}

function toggleMobileMenu() {
    const nav = document.querySelector('.site-nav');
    const button = document.getElementById('mobile-menu-btn');
    if (!nav) return;

    const isOpen = nav.classList.toggle('open');
    if (button) {
        button.setAttribute('aria-expanded', String(isOpen));
    }
}

function showFormMessage(message, isError) {
    const formMessage = document.getElementById('form-message');
    if (!formMessage) return;
    formMessage.style.color = isError ? '#b91c1c' : '#166534';
    formMessage.textContent = message;
}

function parseNumber(value) {
    const n = Number.parseInt(value, 10);
    return Number.isNaN(n) ? 0 : n;
}

function collectRecipePayload(form) {
    const title = form.querySelector('#title')?.value.trim() || '';
    const description = form.querySelector('#description')?.value.trim() || '';
    const prepTime = parseNumber(form.querySelector('#prep_time')?.value || '0');
    const cookTime = parseNumber(form.querySelector('#cook_time')?.value || '0');
    const servings = parseNumber(form.querySelector('#servings')?.value || '1');
    const difficulty = form.querySelector('#difficulty')?.value || '';
    const tagsRaw = form.querySelector('#tags')?.value || '';

    const ingredients = Array.from(form.querySelectorAll('#ingredients-list .ingredient-row'))
        .map((row) => ({
            name: row.querySelector('input[name*=".name"]')?.value.trim() || '',
            amount: row.querySelector('input[name*=".amount"]')?.value.trim() || '',
            unit: row.querySelector('input[name*=".unit"]')?.value.trim() || ''
        }))
        .filter((i) => i.name !== '');

    const instructions = Array.from(form.querySelectorAll('#instructions-list .instruction-row'))
        .map((row, index) => ({
            text: row.querySelector('textarea')?.value.trim() || '',
            position: index + 1
        }))
        .filter((i) => i.text !== '');

    const tags = tagsRaw
        .split(',')
        .map((t) => t.trim())
        .filter((t) => t.length > 0);

    return {
        title,
        description,
        prep_time: prepTime,
        cook_time: cookTime,
        servings,
        difficulty,
        ingredients,
        instructions,
        tags
    };
}

async function handleJSONResponse(response) {
    if (response.status === 204) {
        return null;
    }
    return response.json();
}

function setupAuthForms() {
    const loginForm = document.getElementById('login-form');
    if (loginForm) {
        loginForm.addEventListener('submit', async (event) => {
            event.preventDefault();
            const errorEl = document.getElementById('login-error');
            if (errorEl) {
                errorEl.style.display = 'none';
                errorEl.textContent = '';
            }

            const payload = {
                email: loginForm.querySelector('#email')?.value || '',
                password: loginForm.querySelector('#password')?.value || ''
            };

            const response = await fetch('/api/auth/login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                credentials: 'include',
                body: JSON.stringify(payload)
            });

            if (!response.ok) {
                const text = await response.text();
                if (errorEl) {
                    errorEl.textContent = text || 'Error de autenticación';
                    errorEl.style.display = 'block';
                }
                return;
            }

            window.location.reload();
        });
    }

    const registerForm = document.getElementById('register-form');
    if (registerForm) {
        registerForm.addEventListener('submit', async (event) => {
            event.preventDefault();
            const errorEl = document.getElementById('register-error');
            if (errorEl) {
                errorEl.style.display = 'none';
                errorEl.textContent = '';
            }

            const payload = {
                first_name: registerForm.querySelector('#regFirstName')?.value || '',
                last_name: registerForm.querySelector('#regLastName')?.value || '',
                username: registerForm.querySelector('#regUsername')?.value || '',
                email: registerForm.querySelector('#regEmail')?.value || '',
                password: registerForm.querySelector('#regPassword')?.value || ''
            };

            const response = await fetch('/api/auth/register', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                credentials: 'include',
                body: JSON.stringify(payload)
            });

            if (!response.ok) {
                const text = await response.text();
                if (errorEl) {
                    errorEl.textContent = text || 'No se pudo registrar';
                    errorEl.style.display = 'block';
                }
                return;
            }

            window.location.reload();
        });
    }

    const logoutForm = document.getElementById('logout-form');
    if (logoutForm) {
        logoutForm.addEventListener('submit', async (event) => {
            event.preventDefault();
            await fetch('/api/auth/logout', {
                method: 'POST',
                credentials: 'include'
            });
            window.location.href = '/';
        });
    }
}

function setupCreateRecipeForm() {
    const form = document.getElementById('create-recipe-form');
    if (!form) return;

    form.addEventListener('submit', async (event) => {
        event.preventDefault();
        const payload = collectRecipePayload(form);

        const response = await fetch('/api/recipes', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            credentials: 'include',
            body: JSON.stringify(payload)
        });

        if (!response.ok) {
            const text = await response.text();
            showFormMessage(text || 'No se pudo crear la receta', true);
            return;
        }

        showFormMessage('Receta creada correctamente', false);
        window.setTimeout(() => {
            window.location.href = '/recipes';
        }, 700);
    });
}

function setupEditRecipeForm() {
    const form = document.getElementById('edit-recipe-form');
    if (!form) return;

    const recipeID = form.getAttribute('data-recipe-id');
    if (!recipeID) return;

    async function loadRecipe() {
        const response = await fetch(`/api/recipes/${recipeID}`, { credentials: 'include' });
        if (!response.ok) {
            window.location.href = '/recipes';
            return;
        }

        const recipe = await handleJSONResponse(response);
        if (!recipe || !recipe.is_owner) {
            window.location.href = `/recipes/${recipeID}`;
            return;
        }

        form.querySelector('#title').value = recipe.title || '';
        form.querySelector('#description').value = recipe.description || '';
        form.querySelector('#prep_time').value = recipe.prep_time || 0;
        form.querySelector('#cook_time').value = recipe.cook_time || 0;
        form.querySelector('#servings').value = recipe.servings || 1;
        form.querySelector('#difficulty').value = recipe.difficulty || '';
        form.querySelector('#tags').value = (recipe.tags || []).join(', ');

        const ingredientsList = form.querySelector('#ingredients-list');
        const instructionsList = form.querySelector('#instructions-list');
        if (ingredientsList) ingredientsList.innerHTML = '';
        if (instructionsList) instructionsList.innerHTML = '';
        ingredientCount = 0;
        instructionCount = 0;

        (recipe.ingredients || []).forEach((ingredient) => addIngredient(ingredient));
        (recipe.instructions || []).forEach((instruction) => addInstruction(instruction));

        if (ingredientCount === 0) addIngredient();
        if (instructionCount === 0) addInstruction();
    }

    form.addEventListener('submit', async (event) => {
        event.preventDefault();
        const payload = collectRecipePayload(form);

        const response = await fetch(`/api/recipes/${recipeID}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            credentials: 'include',
            body: JSON.stringify(payload)
        });

        if (!response.ok) {
            const text = await response.text();
            showFormMessage(text || 'No se pudo actualizar la receta', true);
            return;
        }

        window.location.href = `/recipes/${recipeID}`;
    });

    loadRecipe();
}

async function deleteRecipe(recipeID) {
    if (!window.confirm('¿Seguro que quieres eliminar esta receta?')) {
        return;
    }

    const response = await fetch(`/api/recipes/${recipeID}`, {
        method: 'DELETE',
        credentials: 'include'
    });

    if (response.status === 204) {
        window.location.href = '/recipes';
        return;
    }

    if (response.status === 403) {
        showFlashMessage('No tienes permisos para eliminar esta receta', 'error');
        return;
    }

    const text = await response.text();
    showFlashMessage(text || 'No se pudo eliminar la receta', 'error');
}

function showFlashMessage(message, type = 'info') {
    const flashDiv = document.createElement('div');
    flashDiv.className = `fixed top-20 right-4 p-4 rounded-lg shadow-lg z-50 ${
        type === 'error' ? 'bg-red-500 text-white' :
        type === 'success' ? 'bg-green-500 text-white' :
        type === 'warning' ? 'bg-yellow-500 text-white' :
        'bg-blue-500 text-white'
    }`;
    flashDiv.textContent = message;
    document.body.appendChild(flashDiv);

    setTimeout(() => {
        flashDiv.remove();
    }, 4000);
}

document.addEventListener('DOMContentLoaded', () => {
    setupAuthForms();
    setupCreateRecipeForm();
    setupEditRecipeForm();
});