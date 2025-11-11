<template>
  <div class="license-container">
    <div class="header">
      <h1>License管理系统</h1>
    </div>
    
    <div class="license-actions">
      <el-button type="primary" @click="showAddDialog = true">新增License</el-button>
      <el-button type="success" @click="generateFingerprint">测试生成指纹</el-button>
      <el-button @click="refreshData">刷新数据</el-button>
    </div>

    <div class="license-content">
      <div class="content-row">
        <!-- License统计图表 - 小模块 -->
        <div class="license-chart-small">
          <div id="licenseChart"></div>
        </div>
        
        <!-- License列表 -->
        <div class="license-table">
          <div class="table-header">
            <h3>License列表</h3>
          </div>
          <div class="table-container">
            <el-table :data="licenseList" style="width: 100%" table-layout="fixed">
              <el-table-column prop="id" label="ID" width="80" />
              <el-table-column prop="customer" label="客户名称" min-width="120" />
              <el-table-column prop="fingerprint" label="硬件指纹" min-width="150" />
              <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
              <el-table-column prop="activated_at" label="激活时间" min-width="150">
                <template #default="scope">
                  {{ formatDate(scope.row.activated_at) }}
                </template>
              </el-table-column>
              <el-table-column prop="expires_at" label="过期时间" min-width="150">
                <template #default="scope">
                  {{ formatDate(scope.row.expires_at) }}
                </template>
              </el-table-column>
              <el-table-column prop="is_active" label="状态" width="100">
                <template #default="scope">
                  <el-tag :type="scope.row.is_active ? 'success' : 'danger'">
                    {{ scope.row.is_active ? '激活' : '停用' }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column label="操作" width="150">
                <template #default="scope">
                  <el-button size="small" @click="deactivateLicense(scope.row.id)" v-if="scope.row.is_active">
                    停用
                  </el-button>
                  <el-button size="small" type="danger" @click="deleteLicense(scope.row.id)">
                    删除
                  </el-button>
                </template>
              </el-table-column>
            </el-table>
          </div>
          <!-- 分页组件 -->
          <div class="pagination-container">
            <el-pagination
              v-model:current-page="currentPage"
              v-model:page-size="pageSize"
              :page-sizes="[10, 20, 50, 100]"
              :total="total"
              layout="total, sizes, prev, pager, next, jumper"
              prev-text="上一页"
              next-text="下一页"
              :pager-count="7"
              @size-change="handleSizeChange"
              @current-change="handleCurrentChange"
            />
          </div>
        </div>
      </div>
    </div>

    <!-- 新增License对话框 -->
    <el-dialog v-model="showAddDialog" title="新增License" width="600px">
      <el-form :model="newLicense" label-width="80px">
        <el-form-item label="客户名称" required>
          <el-input v-model="newLicense.customer" placeholder="请输入客户名称" />
        </el-form-item>
        <el-form-item label="硬件指纹" required>
          <el-tooltip
            v-model="showFingerprintTooltip"
            :content="fingerprintTooltipContent"
            placement="top"
            :disabled="!showFingerprintTooltip"
          >
            <el-input 
              v-model="newLicense.fingerprint" 
              placeholder="请输入硬件指纹" 
              @blur="handleFingerprintBlur"
              :class="{ 'fingerprint-invalid': showFingerprintValidation && newLicense.fingerprint && !validateFingerprintFormat(newLicense.fingerprint) }"
            />
          </el-tooltip>
          <div style="color: #909399; font-size: 12px; margin-top: 5px;">
            正确格式应为XXXX-XXXX-XXXX-XXXX（4组4位字母或数字）
          </div>
        </el-form-item>
        <el-form-item label="有效期">
          <div class="validity-container">
            <div class="validity-row">
              <div class="validity-item">
                <el-form-item label="天数" prop="validityDays">
                  <el-input-number v-model="newLicense.validityDays" :min="0" :max="3650" style="width: 100%" />
                </el-form-item>
              </div>
              <div class="validity-item">
                <el-form-item label="小时" prop="validityHours">
                  <el-input-number v-model="newLicense.validityHours" :min="0" :max="23" style="width: 100%" />
                </el-form-item>
              </div>
            </div>
            <div class="validity-row">
              <div class="validity-item">
                <el-form-item label="分钟" prop="validityMinutes">
                  <el-input-number v-model="newLicense.validityMinutes" :min="0" :max="59" style="width: 100%" />
                </el-form-item>
              </div>
              <div class="validity-item">
                <el-form-item label="秒" prop="validitySeconds">
                  <el-input-number v-model="newLicense.validitySeconds" :min="0" :max="59" style="width: 100%" />
                </el-form-item>
              </div>
            </div>
          </div>
          <div style="color: #909399; font-size: 12px; margin-top: 5px;">
            至少需要设置一个时间单位
          </div>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="newLicense.description" type="textarea" rows="4" placeholder="请输入描述内容" maxlength="300" show-word-limit />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="showAddDialog = false; showFingerprintValidation = false; showFingerprintTooltip = false">取消</el-button>
          <el-button type="primary" @click="addLicense">确定</el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 指纹生成对话框 -->
    <el-dialog v-model="showFingerprintDialog" title="机器指纹" width="500px">
      <div class="fingerprint-content">
        <div class="fingerprint-label">当前机器指纹：</div>
        <div class="fingerprint-value">
          <el-input v-model="currentFingerprint" readonly>
            <template #append>
              <el-button @click="copyFingerprint" type="primary">复制</el-button>
            </template>
          </el-input>
        </div>
        <div class="fingerprint-tip">
          <el-alert
            title="提示"
            description="此指纹基于当前机器硬件信息生成，可用于License授权。仅用于测试，生产环境请使用真实机器生成。正确格式应为XXXX-XXXX-XXXX-XXXX（4组4位字母或数字）"
            type="info"
            show-icon
            :closable="false"
          />
        </div>
      </div>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="showFingerprintDialog = false">关闭</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import * as echarts from 'echarts'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Check, Close } from '@element-plus/icons-vue'
import axios from 'axios'

// 数据定义
const licenseList = ref([])
const showAddDialog = ref(false)
const showFingerprintDialog = ref(false)
const currentFingerprint = ref('')
const chartInstance = ref(null)
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(10)
const showFingerprintValidation = ref(false)
const showFingerprintTooltip = ref(false)
const fingerprintTooltipContent = ref('')

const newLicense = ref({
  customer: '',
  fingerprint: '',
  validityDays: 0,
  validityHours: 0,
  validityMinutes: 0,
  validitySeconds: 0,
  description: '',
  licenseContent: ''
})

// 日期格式化函数
const formatDate = (dateString) => {
  if (!dateString) return ''
  const date = new Date(dateString)
  if (isNaN(date.getTime())) return dateString
  
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  const seconds = String(date.getSeconds()).padStart(2, '0')
  
  return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`
}

// API基础URL
const API_BASE_URL = 'http://localhost:8080/api'

// 更新图表数据
const updateChart = () => {
  // 销毁旧实例并重新创建，避免type变化的冲突
  if (chartInstance.value) {
    chartInstance.value.dispose()
    chartInstance.value = null
  }
  
  const chartDom = document.getElementById('licenseChart')
  if (!chartDom) return
  
  chartInstance.value = echarts.init(chartDom)
  
  // 计算统计数据
  const totalCount = licenseList.value.length
  const activeCount = licenseList.value.filter(item => item.is_active).length
  const expiredCount = totalCount - activeCount
  
  const option = {
    title: {
      text: 'License统计',
      left: 'center'
    },
    tooltip: {
      trigger: 'item'
    },
    legend: {
      orient: 'vertical',
      left: 'left'
    },
    series: [
      {
        name: 'License状态',
        type: 'pie',
        radius: '50%',
        data: [
          { value: activeCount, name: '激活中' },
          { value: expiredCount, name: '已过期/停用' }
        ],
        emphasis: {
          itemStyle: {
            shadowBlur: 10,
            shadowOffsetX: 0,
            shadowColor: 'rgba(0, 0, 0, 0.5)'
          }
        }
      }
    ]
  }
  
  chartInstance.value.setOption(option)
}

// 获取License列表
const fetchLicenseList = async (page = 1, size = 10) => {
  try {
    const response = await axios.get(`${API_BASE_URL}/license/activations`, {
      params: {
        page,
        size
      }
    })
    
    // 处理后端返回的数据
    // 假设后端可能支持分页或不分页的两种情况
    if (response.data.activations) {
      licenseList.value = response.data.activations || []
      // 如果后端支持分页，会返回总数
      if (response.data.total !== undefined) {
        total.value = response.data.total
      } else {
        total.value = licenseList.value.length
      }
    } else {
      // 兼容不分页的情况
      licenseList.value = response.data || []
      total.value = licenseList.value.length
    }
    
    updateChart()
  } catch (error) {
    console.error('获取License列表失败:', error)
    ElMessage.error('获取License列表失败')
  }
}

// 添加License
const addLicense = async () => {
  try {
    // 验证客户名称必填
    if (!newLicense.value.customer || newLicense.value.customer.trim() === '') {
      ElMessage.error('客户名称不能为空')
      return
    }
    
    // 验证硬件指纹必填
    if (!newLicense.value.fingerprint || newLicense.value.fingerprint.trim() === '') {
      ElMessage.error('硬件指纹不能为空')
      return
    }
    
    // 验证指纹格式
    if (!validateFingerprintFormat(newLicense.value.fingerprint)) {
      showFingerprintValidation.value = true // 确保显示格式校验提示
      showFingerprintTooltip.value = true // 显示tooltip
      fingerprintTooltipContent.value = '硬件指纹格式不正确，应为XXXX-XXXX-XXXX-XXXX格式（4组4位字母或数字）'
      ElMessage.error('硬件指纹格式不正确，请按照提示修改')
      return
    }
    
    // 验证至少有一个时间单位被设置
    if (newLicense.value.validityDays === 0 && 
        newLicense.value.validityHours === 0 && 
        newLicense.value.validityMinutes === 0 && 
        newLicense.value.validitySeconds === 0) {
      ElMessage.error('请至少设置一个时间单位')
      return
    }
    
    // 首先激活License
    const requestData = {
      customer: newLicense.value.customer,
      fingerprint: newLicense.value.fingerprint,
      description: newLicense.value.description,
      validityDays: newLicense.value.validityDays,
      validityHours: newLicense.value.validityHours,
      validityMinutes: newLicense.value.validityMinutes,
      validitySeconds: newLicense.value.validitySeconds
    }
    
    const activateResponse = await axios.post(`${API_BASE_URL}/license/activate`, requestData)
    
    if (activateResponse.data.success) {
      ElMessage.success('License添加成功')
      showAddDialog.value = false
      showFingerprintValidation.value = false // 重置指纹格式校验提示状态
      showFingerprintTooltip.value = false // 重置tooltip状态
      // 重置表单
      newLicense.value = {
        customer: '',
        fingerprint: '',
        validityDays: 0,
        validityHours: 0,
        validityMinutes: 0,
        validitySeconds: 0,
        description: '',
        licenseContent: ''
      }
      // 刷新列表
      fetchLicenseList()
    } else {
      ElMessage.error(activateResponse.data.message || 'License添加失败')
    }
  } catch (error) {
    console.error('添加License失败:', error)
    ElMessage.error('添加License失败')
  }
}

// 停用License
const deactivateLicense = async (id) => {
  try {
    await ElMessageBox.confirm('确定要停用这个License吗?', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    // 调用后端API停用License
    const response = await axios.put(`${API_BASE_URL}/license/activations/${id}/deactivate`)
    
    if (response.data.success) {
      ElMessage.success('License停用成功')
      // 刷新列表
      fetchLicenseList(currentPage.value, pageSize.value)
    } else {
      ElMessage.error('License停用失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('停用License时出错: ' + (error.response?.data?.error || error.message))
    }
  }
}

// 删除License
const deleteLicense = async (id) => {
  try {
    await ElMessageBox.confirm('确定要删除这个License吗?', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    // 调用后端API删除License
    const response = await axios.delete(`${API_BASE_URL}/license/activations/${id}`)
    
    if (response.data.success) {
      ElMessage.success('License删除成功')
      // 刷新列表
      fetchLicenseList(currentPage.value, pageSize.value)
    } else {
      ElMessage.error('License删除失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除License时出错: ' + (error.response?.data?.error || error.message))
    }
  }
}

// 刷新数据
const refreshData = () => {
  fetchLicenseList(currentPage.value, pageSize.value)
  ElMessage.success('数据已刷新')
}

// 指纹格式校验函数
const validateFingerprintFormat = (fingerprint) => {
  // 指纹格式应为：XXXX-XXXX-XXXX-XXXX，其中X为字母或数字
  const pattern = /^[A-Z0-9]{4}-[A-Z0-9]{4}-[A-Z0-9]{4}-[A-Z0-9]{4}$/i
  return pattern.test(fingerprint)
}

// 处理指纹输入框失去焦点事件
const handleFingerprintBlur = () => {
  showFingerprintValidation.value = true
  
  if (newLicense.value.fingerprint) {
    if (!validateFingerprintFormat(newLicense.value.fingerprint)) {
      showFingerprintTooltip.value = true
      fingerprintTooltipContent.value = '硬件指纹格式不正确，应为XXXX-XXXX-XXXX-XXXX格式（4组4位字母或数字）'
    } else {
      showFingerprintTooltip.value = false
    }
  } else {
    showFingerprintTooltip.value = false
  }
}

// 生成指纹
const generateFingerprint = async () => {
  try {
    const response = await axios.get(`${API_BASE_URL}/system/fingerprint`)
    if (response.data && response.data.fingerprint) {
      currentFingerprint.value = response.data.fingerprint
      showFingerprintDialog.value = true
    } else {
      ElMessage.error('获取指纹失败')
    }
  } catch (error) {
    console.error('获取指纹失败:', error)
    ElMessage.error('获取指纹失败')
  }
}

// 复制指纹
const copyFingerprint = () => {
  if (navigator.clipboard) {
    navigator.clipboard.writeText(currentFingerprint.value)
      .then(() => {
        ElMessage.success('指纹已复制到剪贴板')
      })
      .catch(err => {
        console.error('复制失败:', err)
        ElMessage.error('复制失败')
      })
  } else {
    // 降级方案
    const textArea = document.createElement('textarea')
    textArea.value = currentFingerprint.value
    document.body.appendChild(textArea)
    textArea.focus()
    textArea.select()
    try {
      document.execCommand('copy')
      ElMessage.success('指纹已复制到剪贴板')
    } catch (err) {
      console.error('复制失败:', err)
      ElMessage.error('复制失败')
    }
    document.body.removeChild(textArea)
  }
}



// 每页条数变化处理
const handleSizeChange = (size) => {
  pageSize.value = size
  currentPage.value = 1
  fetchLicenseList(currentPage.value, pageSize.value)
}

// 当前页变化处理
const handleCurrentChange = (page) => {
  currentPage.value = page
  fetchLicenseList(currentPage.value, pageSize.value)
}

// 组件挂载时初始化
onMounted(() => {
  fetchLicenseList(currentPage.value, pageSize.value)
  
  // 窗口大小变化时重新渲染图表
  window.addEventListener('resize', () => {
    // 使用setTimeout避免在main process期间调用resize
    setTimeout(() => {
      if (chartInstance.value && !chartInstance.value.isDisposed()) {
        chartInstance.value.resize()
      }
    }, 0)
  })
})
</script>

<style scoped>
.license-container {
  padding: 20px;
  width: 100%;
  margin: 0;
  height: 100vh; /* 改为固定高度，铺满整个视口 */
  display: flex;
  flex-direction: column;
  background-color: #f5f7fa;
  box-sizing: border-box; /* 确保padding包含在高度内 */
}

.header {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 30px 0;
  text-align: center;
  border-radius: 8px;
  margin-bottom: 30px;
  box-shadow: 0 4px 15px rgba(0, 0, 0, 0.1);
}

h1 {
  margin: 0;
  font-size: 2.2rem;
  font-weight: 600;
  letter-spacing: -0.5px;
}

.license-actions {
  margin-bottom: 20px;
  display: flex;
  gap: 15px;
  padding: 20px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.08);
}

.license-content {
  display: flex;
  flex-direction: column;
  gap: 24px;
  flex: 1; /* 填充剩余空间 */
}

.content-row {
  display: flex;
  gap: 24px;
  align-items: stretch; /* 改为stretch使子元素高度一致 */
  flex: 1; /* 填充剩余空间 */
}

.license-chart-small {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.08);
  padding: 20px;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
  width: 350px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  min-height: 0; /* 确保flex子元素可以缩小 */
}

.license-chart-small:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.12);
}

.license-chart {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.08);
  padding: 24px;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.license-chart:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.12);
}

.license-table {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.08);
  padding: 24px;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0; /* 确保flex子元素可以缩小 */
}

.table-container {
  flex: 1; /* 表格容器填充剩余空间 */
  overflow: auto; /* 添加滚动条以防内容过多 */
  min-height: 0; /* 确保flex子元素可以缩小 */
}

/* 表格样式调整，确保每条记录宽度一致 */
.table-container :deep(.el-table) {
  width: 100% !important; /* 确保表格宽度为100% */
}

.table-container :deep(.el-table__body-wrapper) {
  width: 100% !important;
}

.table-container :deep(.el-table__header-wrapper) {
  width: 100% !important;
}

.table-container :deep(.el-table__body) {
  width: 100% !important;
}

.table-container :deep(.el-table__header) {
  width: 100% !important;
}

.table-container :deep(.el-table__row) {
  width: 100% !important;
}

.table-container :deep(.el-table__cell) {
  padding: 12px 0; /* 调整单元格内边距 */
}

.pagination-container {
  margin-top: 20px;
  display: flex;
  justify-content: center;
  padding: 20px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.08);
  flex-shrink: 0; /* 防止分页组件被压缩 */
}

.table-header {
  margin-bottom: 20px;
  border-bottom: 2px solid #f0f0f0;
  padding-bottom: 10px;
}

.table-header h3 {
  margin: 0;
  color: #333;
  font-size: 1.2rem;
  font-weight: 600;
}

.license-table:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.12);
}

#licenseChart {
  width: 100%;
  flex: 1; /* 使图表填充剩余空间 */
  min-height: 300px; /* 设置最小高度 */
}

/* 指纹输入框校验不正确时的样式 */
.fingerprint-invalid .el-input__wrapper {
  border-color: var(--el-color-danger) !important;
  box-shadow: 0 0 0 1px var(--el-color-danger) inset;
}

.fingerprint-invalid .el-input__wrapper:hover {
  border-color: var(--el-color-danger) !important;
}

.fingerprint-invalid .el-input__wrapper.is-focus {
  border-color: var(--el-color-danger) !important;
  box-shadow: 0 0 0 1px var(--el-color-danger) inset;
}

/* 确保tooltip样式正确 */
.el-popper.is-dark {
  background-color: var(--el-color-danger) !important;
  color: #fff !important;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .license-container {
    padding: 15px;
  }
  
  .header {
    padding: 20px 0;
  }
  
  h1 {
    font-size: 1.8rem;
  }
  
  .license-actions {
    flex-direction: column;
    gap: 10px;
    padding: 15px;
  }
  
  .license-chart,
  .license-table {
    padding: 15px;
  }
  
  #licenseChart {
    height: 200px;
  }
  
  .pagination-container {
    justify-content: center;
  }
}

/* 有效期输入区域样式 */
.validity-container {
  width: 100%;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  padding: 12px;
  background-color: #f9f9f9;
  box-sizing: border-box; /* 确保padding不会增加总宽度 */
  overflow: hidden; /* 防止内容溢出 */
}

.validity-row {
  margin-bottom: 10px;
  display: flex;
  gap: 10px; /* 使用gap代替el-col的gutter */
  width: 100%;
  box-sizing: border-box;
}

.validity-row:last-child {
  margin-bottom: 0;
}

.validity-item {
  margin-bottom: 0;
  flex: 1; /* 让每个输入框占据相等的空间 */
  min-width: 0; /* 允许缩小 */
  box-sizing: border-box;
  overflow: hidden; /* 防止内容溢出 */
}

.validity-item :deep(.el-form-item) {
  margin-bottom: 0;
  width: 100%;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
}

.validity-item :deep(.el-form-item__label) {
  font-weight: 500;
  color: #606266;
  line-height: 1.2;
  padding-bottom: 4px;
  width: 100%;
  box-sizing: border-box;
  text-align: left;
  flex-shrink: 0;
}

.validity-item :deep(.el-form-item__content) {
  width: 100%;
  box-sizing: border-box;
  flex: 1;
}

.validity-item :deep(.el-input-number) {
  width: 100%;
  box-sizing: border-box;
}

.validity-item :deep(.el-input-number .el-input__inner) {
  text-align: center;
  box-sizing: border-box;
}

/* 指纹对话框样式 */
.fingerprint-content {
  padding: 10px 0;
}

.fingerprint-label {
  margin-bottom: 10px;
  font-weight: 500;
  color: #303133;
}

.fingerprint-value {
  margin-bottom: 15px;
}

.fingerprint-format {
  margin-bottom: 15px;
  padding: 8px 12px;
  border-radius: 4px;
  font-size: 14px;
  display: flex;
  align-items: center;
}

.format-valid {
  background-color: #f0f9ff;
  border: 1px solid #b3d8ff;
  color: #409eff;
}

.format-invalid {
  background-color: #fef0f0;
  border: 1px solid #fbc4c4;
  color: #f56c6c;
}

.fingerprint-tip {
  margin-top: 15px;
}
</style>