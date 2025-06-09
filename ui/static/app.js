const balance = document.getElementById('balance');
const transactionList = document.getElementById('transaction-list');
const form = document.getElementById('transaction-form');
const description = document.getElementById('description');
const amount = document.getElementById('amount');

const API_URL = '/api/transactions';

// Fetch transactions from the backend
async function getTransactions() {
    try {
        const res = await fetch(API_URL);
        const data = await res.json();
        
        // Clear the list before adding new items
        transactionList.innerHTML = '';

        if (data) {
            data.forEach(addTransactionDOM);
        }
        updateBalance();
    } catch (error) {
        console.error('Error fetching transactions:', error);
    }
}

// Add a new transaction via the backend
async function addTransaction(e) {
    e.preventDefault();

    if (description.value.trim() === '' || amount.value.trim() === '') {
        alert('Please add a description and amount');
        return;
    }

    const newTransaction = {
        description: description.value,
        amount: +amount.value,
    };

    try {
        const res = await fetch(API_URL, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(newTransaction),
        });

        const data = await res.json();
        addTransactionDOM(data);
        updateBalance();

        description.value = '';
        amount.value = '';
    } catch (error) {
        console.error('Error adding transaction:', error);
    }
}


// Add transaction to the DOM list
function addTransactionDOM(transaction) {
    const sign = transaction.amount < 0 ? '-' : '+';
    const item = document.createElement('tr');

    const amountClass = transaction.amount < 0 ? 'expense' : 'income';

    item.innerHTML = `
        <td>${transaction.description}</td>
        <td class="${amountClass}">${sign}₹${Math.abs(transaction.amount).toFixed(2)}</td>
        <td>${new Date(transaction.date).toLocaleDateString()}</td>
    `;
    transactionList.appendChild(item);
}


// Update the balance total
function updateBalance() {
    const rows = transactionList.querySelectorAll('tr');
    let total = 0;

    rows.forEach(row => {
        const amountCell = row.querySelector('td:nth-child(2)');
        if (amountCell) {
            // Remove currency symbol, sign, and parse the number
            const amountValue = parseFloat(amountCell.textContent.replace(/[₹+-]/g, ''));
            if(amountCell.classList.contains('expense')){
                total -= amountValue;
            } else {
                total += amountValue;
            }
        }
    });

    balance.innerText = `₹${total.toFixed(2)}`;
}


// Event listeners
form.addEventListener('submit', addTransaction);

// Initial load
getTransactions();
