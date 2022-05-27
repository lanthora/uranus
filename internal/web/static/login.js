const { Button, TextField, FormControl, Container, Box } = window['MaterialUI'];

const USERNAME = "用户名";
const PASSWORD = "密码";
const LOGIN = "登录";

function LoginForm() {
        return (
                <FormControl>
                        <TextField id="username" label={USERNAME} size="small" />
                        <TextField id="password" label={PASSWORD} size="small" type="password" margin="dense" />
                        <Button variant="contained"> {LOGIN} </Button>
                </FormControl>
        )
}

function Login() {
        return (
                <span className="login-center">
                        <LoginForm />
                </span>
        )
}

const root = ReactDOM.createRoot(document.getElementById('root'));
root.render(<Login />);
