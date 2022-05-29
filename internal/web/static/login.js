const { Button, Snackbar, TextField, FormControl, Alert } = window['MaterialUI'];

const LOGIN_TITLE = "登录"
const LOGIN_USERNAME_TEXEFIELD = "用户名";
const LOGIN_PASSWORD_TEXEFIELD = "密码";
const LOGIN_BUTTON_TEXT = "登录";
const LOGIN_FAILED_MSG = "登录失败"
const LOGIN_FAILED_MSG_TIMEOUT = 3000


function Login() {

        React.useEffect(() => {
                document.title = LOGIN_TITLE;
        });

        const [open, setOpen] = React.useState(false);

        const username = React.useRef(null);
        const password = React.useRef(null);

        const openErrorNotify = () => {
                setOpen(true);
        };

        const closeErrorNotify = (event, reason) => {
                if (reason === 'clickaway') {
                        return;
                }
                setOpen(false);
        };

        function userLogin() {
                // TODO: 检查用户名密码非空

                axios.post('/user/login', {
                        "username": username.current.value,
                        "password": password.current.value,
                }).then(function (response) {
                        if (response.status == 200) {
                                window.location.replace("/");
                        } else {
                                openErrorNotify();
                        }
                }).catch(function (error) {
                        if (error.response && error.response.status === 401) {
                                openErrorNotify();
                        }
                });
        }


        return (
                <span>
                        <div className="screen-center">
                                <FormControl>
                                        <TextField
                                                id="username"
                                                label={LOGIN_USERNAME_TEXEFIELD}
                                                size="small"
                                                inputRef={username} />
                                        <TextField
                                                id="password"
                                                label={LOGIN_PASSWORD_TEXEFIELD}
                                                size="small"
                                                inputRef={password}
                                                type="password"
                                                margin="dense" />
                                        <Button
                                                variant="contained"
                                                onClick={userLogin}>
                                                {LOGIN_BUTTON_TEXT}
                                        </Button>
                                </FormControl>
                        </div>
                        <Snackbar
                                anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
                                open={open}
                                autoHideDuration={LOGIN_FAILED_MSG_TIMEOUT}
                                onClose={closeErrorNotify}>
                                <Alert
                                        onClose={closeErrorNotify}
                                        severity="error">
                                        {LOGIN_FAILED_MSG}
                                </Alert>
                        </Snackbar>
                </span>
        )
}

const root = ReactDOM.createRoot(document.getElementById('root'));
root.render(<Login />);
