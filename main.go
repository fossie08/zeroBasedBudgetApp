package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "os"
    "strconv"

    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/app"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/widget"
)

type Income struct {
    Source string  `json:"source"`
    Amount float64 `json:"amount"`
}

type Expense struct {
    Category       string  `json:"category"`
    BudgetedAmount float64 `json:"budgeted_amount"`
    ActualAmount   float64 `json:"actual_amount"`
}

type Budget struct {
    Month     string    `json:"month"`
    Incomes   []Income  `json:"incomes"`
    Expenses  []Expense `json:"expenses"`
    Savings   float64   `json:"savings"`
    Balance   float64   `json:"balance"`
}

type BudgetApp struct {
    FileName string
    Budgets  []Budget
}

// Load data from JSON file
func (app *BudgetApp) LoadData() {
    file, err := ioutil.ReadFile(app.FileName)
    if err != nil {
        if os.IsNotExist(err) {
            app.Budgets = []Budget{}
            return
        }
        panic(err)
    }
    if err := json.Unmarshal(file, &app.Budgets); err != nil {
        panic(err)
    }
}

// Save data to JSON file
func (app *BudgetApp) SaveData() {
    data, err := json.MarshalIndent(app.Budgets, "", "  ")
    if err != nil {
        panic(err)
    }
    if err := ioutil.WriteFile(app.FileName, data, 0644); err != nil {
        panic(err)
    }
}

// Add a new budget for the specified month
func (app *BudgetApp) AddBudget(month string) {
    for _, budget := range app.Budgets {
        if budget.Month == month {
            return
        }
    }
    newBudget := Budget{
        Month:    month,
        Incomes:  []Income{},
        Expenses: []Expense{},
        Savings:  0,
        Balance:  0,
    }
    app.Budgets = append(app.Budgets, newBudget)
}

// Recalculate the balance for a specific month
func (app *BudgetApp) Recalculate(month string) {
    for i, budget := range app.Budgets {
        if budget.Month == month {
            totalIncome := 0.0
            for _, income := range budget.Incomes {
                totalIncome += income.Amount
            }
            totalExpenses := 0.0
            for _, expense := range budget.Expenses {
                totalExpenses += expense.ActualAmount
            }
            app.Budgets[i].Balance = totalIncome - totalExpenses - budget.Savings
            return
        }
    }
}

func main() {
    budgetApp := BudgetApp{FileName: "budget.json"}
    budgetApp.LoadData()

    fyneApp := app.New()
    mainWindow := fyneApp.NewWindow("Zero-Based Budgeting App")

    monthEntry := widget.NewEntry()
    monthEntry.SetPlaceHolder("Enter month (YYYY-MM)")
    incomeSourceEntry := widget.NewEntry()
    incomeSourceEntry.SetPlaceHolder("Enter income source")
    incomeAmountEntry := widget.NewEntry()
    incomeAmountEntry.SetPlaceHolder("Enter income amount")
    expenseCategoryEntry := widget.NewEntry()
    expenseCategoryEntry.SetPlaceHolder("Enter expense category")
    expenseBudgetedEntry := widget.NewEntry()
    expenseBudgetedEntry.SetPlaceHolder("Enter budgeted amount")
    expenseActualEntry := widget.NewEntry()
    expenseActualEntry.SetPlaceHolder("Enter actual amount")
    savingsEntry := widget.NewEntry()
    savingsEntry.SetPlaceHolder("Enter savings amount")

    statusLabel := widget.NewLabel("")

    addBudgetButton := widget.NewButton("Add Budget", func() {
        month := monthEntry.Text
        if month == "" {
            statusLabel.SetText("Month cannot be empty")
            return
        }
        budgetApp.AddBudget(month)
        budgetApp.SaveData()
        statusLabel.SetText("Budget added for " + month)
    })

    addIncomeButton := widget.NewButton("Add Income", func() {
        month := monthEntry.Text
        source := incomeSourceEntry.Text
        amount := incomeAmountEntry.Text
        if month == "" || source == "" || amount == "" {
            statusLabel.SetText("Please fill in all income fields")
            return
        }
        amountValue, err := strconv.ParseFloat(amount, 64)
        if err != nil {
            statusLabel.SetText("Invalid income amount")
            return
        }
        for i, budget := range budgetApp.Budgets {
            if budget.Month == month {
                budgetApp.Budgets[i].Incomes = append(budgetApp.Budgets[i].Incomes, Income{Source: source, Amount: amountValue})
                budgetApp.Recalculate(month)
                budgetApp.SaveData()
                statusLabel.SetText("Income added to " + month)
                return
            }
        }
        statusLabel.SetText("Budget for this month does not exist")
    })

    addExpenseButton := widget.NewButton("Add Expense", func() {
        month := monthEntry.Text
        category := expenseCategoryEntry.Text
        budgeted := expenseBudgetedEntry.Text
        actual := expenseActualEntry.Text
        if month == "" || category == "" || budgeted == "" || actual == "" {
            statusLabel.SetText("Please fill in all expense fields")
            return
        }
        budgetedValue, err := strconv.ParseFloat(budgeted, 64)
        if err != nil {
            statusLabel.SetText("Invalid budgeted amount")
            return
        }
        actualValue, err := strconv.ParseFloat(actual, 64)
        if err != nil {
            statusLabel.SetText("Invalid actual amount")
            return
        }
        for i, budget := range budgetApp.Budgets {
            if budget.Month == month {
                budgetApp.Budgets[i].Expenses = append(budgetApp.Budgets[i].Expenses, Expense{Category: category, BudgetedAmount: budgetedValue, ActualAmount: actualValue})
                budgetApp.Recalculate(month)
                budgetApp.SaveData()
                statusLabel.SetText("Expense added to " + month)
                return
            }
        }
        statusLabel.SetText("Budget for this month does not exist")
    })

    viewBudgetButton := widget.NewButton("View Budget", func() {
        month := monthEntry.Text
        if month == "" {
            statusLabel.SetText("Month cannot be empty")
            return
        }
        for _, budget := range budgetApp.Budgets {
            if budget.Month == month {
                details := fmt.Sprintf("Incomes:\n")
                for _, income := range budget.Incomes {
                    details += fmt.Sprintf("  %s: %.2f\n", income.Source, income.Amount)
                }
                details += fmt.Sprintf("Expenses:\n")
                for _, expense := range budget.Expenses {
                    details += fmt.Sprintf("  %s (Budgeted: %.2f, Actual: %.2f)\n", expense.Category, expense.BudgetedAmount, expense.ActualAmount)
                }
                details += fmt.Sprintf("Savings: %.2f\nBalance: %.2f\n", budget.Savings, budget.Balance)
                statusLabel.SetText(details)
                return
            }
        }
        statusLabel.SetText("Budget for this month does not exist")
    })

    mainWindow.SetContent(container.NewVBox(
        widget.NewLabel("Zero-Based Budgeting App"),
        monthEntry,
		widget.NewAccordion(
            widget.NewAccordionItem("Income", container.NewVBox(
                widget.NewLabel("Income Source"),
                incomeSourceEntry,
                widget.NewLabel("Income Amount"),
                incomeAmountEntry,
                addIncomeButton,
            )),
            widget.NewAccordionItem("Expense", container.NewVBox(
                widget.NewLabel("Expense Category"),
                expenseCategoryEntry,
                widget.NewLabel("Budgeted Amount"),
                expenseBudgetedEntry,
                widget.NewLabel("Actual Amount"),
                expenseActualEntry,
                addExpenseButton,
            )),
            widget.NewAccordionItem("Savings", container.NewVBox(
                widget.NewLabel("Savings Amount"),
                savingsEntry,
            )),
        ),
        addBudgetButton,
        viewBudgetButton,
        statusLabel,
    ))

    mainWindow.Resize(fyne.NewSize(600, 400))
    mainWindow.ShowAndRun()
}